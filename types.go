package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dustin/go-humanize"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/mattn/go-runewidth"
)

type Channel struct {
	File   string
	Worker int
}

type Game struct {
	File      string
	Duration  int64
	Map       string
	Mode      string
	Timestamp pgtype.Timestamp
	Type      string
	Teams     []*Team
}

type Item struct {
	ModTime time.Time
	Path    string
}

type Model struct {
	Channels struct {
		Input  chan string
		Output chan Channel
	}
	Progress struct {
		Completed int
		Remaining int
		Time      time.Time
		Total     int
	}
	Table map[int]string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case Channel:
		if message.File != "" {
			m.Progress.Completed++
			m.Progress.Remaining--
		}
		m.Table[message.Worker] = message.File
		if m.Progress.Completed == m.Progress.Total {
			return m, tea.Quit
		}
	case tea.KeyMsg:
		switch message.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	return m.View1() + m.View2()
}

func (m Model) View1() string {
	b := &strings.Builder{}

	w := table.NewWriter()

	w.AppendHeader(table.Row{"Worker", "File"})
	w.SetOutputMirror(b)
	w.SetStyle(table.StyleLight)

	w.Style().Format.Header = text.FormatDefault

	for worker := 1; worker <= len(m.Table); worker++ {
		file := m.Table[worker]
		if runewidth.StringWidth(file) > 64 {
			a := file[:16]
			z := file[len(file)-48:]
			file = a + "..." + z
		}
		w.AppendRow(table.Row{fmt.Sprintf("%02d", worker), file})
	}

	w.Render()

	return b.String()
}

func (m Model) View2() string {
	eta := ""
	if m.Progress.Completed > 0 {
		since := time.Since(m.Progress.Time)
		frequency := float64(m.Progress.Completed) / since.Seconds()
		if frequency > 0 {
			eta = (time.Duration(float64(m.Progress.Remaining)/frequency) * time.Second).Round(time.Second).String()
		}
	}

	b := &strings.Builder{}

	w := table.NewWriter()

	w.SetOutputMirror(b)
	w.SetStyle(table.StyleLight)
	w.SetColumnConfigs([]table.ColumnConfig{{Number: 1, Align: text.AlignLeft}, {Number: 2, Align: text.AlignRight}})

	w.Style().Format.Header = text.FormatDefault

	w.AppendRow(table.Row{"Total", humanize.Comma(int64(m.Progress.Total))})
	w.AppendRow(table.Row{"Completed", humanize.Comma(int64(m.Progress.Completed))})
	w.AppendRow(table.Row{"Remaining", humanize.Comma(int64(m.Progress.Remaining))})
	w.AppendSeparator()
	w.AppendRow(table.Row{"Percentage", fmt.Sprintf("%d%%", m.Progress.Completed*100/m.Progress.Total)})
	w.AppendRow(table.Row{"Elapsed", time.Since(m.Progress.Time).Round(time.Second).String()})
	w.AppendRow(table.Row{"ETA", eta})

	w.Render()

	return b.String()
}

func NewModel(total int) Model {
	m := Model{}
	m.Channels.Input = make(chan string)
	m.Channels.Output = make(chan Channel)
	m.Progress.Completed = 0
	m.Progress.Remaining = total
	m.Progress.Time = time.Now()
	m.Progress.Total = total
	m.Table = make(map[int]string)
	return m
}

type Player struct {
	Team          int64
	Number        int64
	APM           int64
	Color         string
	Control       string
	MMR           int64
	Name          string
	Observe       string
	RacesAssigned string
	RacesSelected string
	Result        string
	Messages      []*Message
	Stats         []*Stat
	Units         []*Unit
}

type Message struct {
	Player      int64
	Time        int64
	RecipientID int64 // 0 = All, 2 = Allies, 4 = Observers
	String      string
}

type Stat struct {
	Player                           int64
	Time                             int64
	FoodMade                         int64
	FoodUsed                         int64
	MineralsCollectionRate           int64
	MineralsCurrent                  int64
	MineralsFriendlyFireArmy         int64
	MineralsFriendlyFireEconomy      int64
	MineralsFriendlyFireTechnology   int64
	MineralsKilledArmy               int64
	MineralsKilledEconomy            int64
	MineralsKilledTechnology         int64
	MineralsLostArmy                 int64
	MineralsLostEconomy              int64
	MineralsLostTechnology           int64
	MineralsUsedActiveForces         int64
	MineralsUsedCurrentArmy          int64
	MineralsUsedCurrentEconomy       int64
	MineralsUsedCurrentTechnology    int64
	MineralsUsedInProgressArmy       int64
	MineralsUsedInProgressEconomy    int64
	MineralsUsedInProgressTechnology int64
	VespeneCollectionRate            int64
	VespeneCurrent                   int64
	VespeneFriendlyFireArmy          int64
	VespeneFriendlyFireEconomy       int64
	VespeneFriendlyFireTechnology    int64
	VespeneKilledArmy                int64
	VespeneKilledEconomy             int64
	VespeneKilledTechnology          int64
	VespeneLostArmy                  int64
	VespeneLostEconomy               int64
	VespeneLostTechnology            int64
	VespeneUsedActiveForces          int64
	VespeneUsedCurrentArmy           int64
	VespeneUsedCurrentEconomy        int64
	VespeneUsedCurrentTechnology     int64
	VespeneUsedInProgressArmy        int64
	VespeneUsedInProgressEconomy     int64
	VespeneUsedInProgressTechnology  int64
	WorkersActiveCount               int64
}

type Team struct {
	Number  int64
	Result  string
	Players []*Player
}

type Unit struct {
	Player int64
	Time   int64
	Action string
	Name   string
	X      int64
	Y      int64
}
