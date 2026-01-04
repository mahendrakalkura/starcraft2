package main

import (
	"encoding/json"
	"fmt"
)

func dump(i any) string { // nolint
	mi, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		panic(fmt.Errorf("json.MarshalIndent(): %w", err))
	}

	return string(mi)
}

func unused(i any) {} // nolint
