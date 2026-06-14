package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"main/models"
)

const (
	statusDuplicate = "duplicate"
	statusFailed    = "failed"
	statusImported  = "imported"
	statusSkippedAI = "skipped_ai"
)

func buildFiles(paths []string) ([]string, error) {
	files := []string{}

	for _, path := range paths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !strings.HasSuffix(path, ".SC2Replay") {
				return nil
			}

			files = append(files, path)

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("filepath.Walk(%s): %w", path, err)
		}
	}

	return files, nil
}

func ingest(ctx context.Context, application *Application, force bool) error {
	if len(application.Settings.Replays) == 0 {
		return fmt.Errorf("REPLAYS environment variable is required for ingest")
	}

	sidecar, err := locateSidecar()
	if err != nil {
		return fmt.Errorf("locateSidecar(): %w", err)
	}

	if force {
		err = application.Queries.FilesDeleteAll(ctx)
		if err != nil {
			return fmt.Errorf("application.Queries.FilesDeleteAll(): %w", err)
		}
		fmt.Println("force: cleared all imported data")
	}

	files, err := buildFiles(application.Settings.Replays)
	if err != nil {
		return fmt.Errorf("buildFiles(): %w", err)
	}

	known, err := application.Queries.FilesSelectPaths(ctx)
	if err != nil {
		return fmt.Errorf("application.Queries.FilesSelectPaths(): %w", err)
	}
	knownSet := make(map[string]bool, len(known))
	for _, path := range known {
		knownSet[path] = true
	}

	pending := []string{}
	for _, path := range files {
		if !knownSet[path] {
			pending = append(pending, path)
		}
	}

	fmt.Printf("found %d replay files, %d already processed, %d pending\n", len(files), len(files)-len(pending), len(pending))
	if len(pending) == 0 {
		return nil
	}

	input := make(chan string)
	go func() {
		defer close(input)
		for _, path := range pending {
			select {
			case <-ctx.Done():
				return
			case input <- path:
			}
		}
	}()

	completed := atomic.Int64{}
	counts := sync.Map{}

	wg := sync.WaitGroup{}
	for range application.Settings.Workers {
		wg.Go(func() {
			for path := range input {
				status, e := processFile(ctx, application, sidecar, path)
				if e != nil {
					if ctx.Err() != nil {
						return
					}
					status = fmt.Sprintf("error (%v)", e)
				}
				count, _ := counts.LoadOrStore(status, &atomic.Int64{})
				count.(*atomic.Int64).Add(1)
				fmt.Printf("%d/%d %s %s\n", completed.Add(1), len(pending), status, path)
			}
		})
	}
	wg.Wait()

	fmt.Println("summary:")
	counts.Range(func(status any, count any) bool {
		fmt.Printf("  %s: %d\n", status, count.(*atomic.Int64).Load())
		return true
	})

	return nil
}

func processFile(ctx context.Context, application *Application, sidecar string, path string) (string, error) {
	replay, err := runSidecar(ctx, sidecar, path)
	if err != nil {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		e := recordFile(ctx, application, path, statusFailed, "")
		if e != nil {
			return "", e
		}
		return statusFailed, nil
	}

	if replay.Skip == "ai" {
		err = recordFile(ctx, application, path, statusSkippedAI, replay.ParserVersion)
		if err != nil {
			return "", err
		}
		return statusSkippedAI, nil
	}

	resolveResults(replay, application.Settings.Players)

	print := fingerprint(replay)

	exists, err := application.Queries.GamesFingerprintExists(ctx, print)
	if err != nil {
		return "", fmt.Errorf("application.Queries.GamesFingerprintExists(): %w", err)
	}
	if exists {
		err = recordFile(ctx, application, path, statusDuplicate, replay.ParserVersion)
		if err != nil {
			return "", err
		}
		return statusDuplicate, nil
	}

	err = insertReplay(ctx, application, path, replay, print)
	if err != nil {
		if isFingerprintConflict(err) {
			e := recordFile(ctx, application, path, statusDuplicate, replay.ParserVersion)
			if e != nil {
				return "", e
			}
			return statusDuplicate, nil
		}
		return "", fmt.Errorf("insertReplay(): %w", err)
	}

	return statusImported, nil
}

func recordFile(ctx context.Context, application *Application, path string, status string, parserVersion string) error {
	fiop := models.FilesInsertOneParams{
		Path:          path,
		ParserVersion: parserVersion,
		Status:        status,
	}
	_, err := application.Queries.FilesInsertOne(ctx, fiop)
	if err != nil {
		return fmt.Errorf("application.Queries.FilesInsertOne(): %w", err)
	}

	return nil
}
