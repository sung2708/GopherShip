package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sungp/gophership/pkg/protocol"
)

// Dashboard holds the TUI state for gs-ctl top.
type Dashboard struct {
	app    *tview.Application
	pages  *tview.Pages
	status *tview.TextView
	info   *tview.Table
}

// NewDashboard initializes the terminal UI components.
func NewDashboard() *Dashboard {
	app := tview.NewApplication()

	// Status Header
	status := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[white:blue:b] GopherShip Somatic Dashboard [white:-:-]")

	// Metrics Table
	table := tview.NewTable().
		SetBorders(true).
		SetSelectable(false, false)

	table.SetCell(0, 0, tview.NewTableCell("Metric").SetTextColor(tcell.ColorYellow).SetAttributes(tcell.AttrBold))
	table.SetCell(0, 1, tview.NewTableCell("Value").SetTextColor(tcell.ColorYellow).SetAttributes(tcell.AttrBold))

	table.SetCell(1, 0, tview.NewTableCell("Somatic Zone"))
	table.SetCell(2, 0, tview.NewTableCell("Pressure Score"))
	table.SetCell(3, 0, tview.NewTableCell("Memory Usage"))
	table.SetCell(4, 0, tview.NewTableCell("Heap Objects"))
	table.SetCell(5, 0, tview.NewTableCell("Goroutine Count"))

	return &Dashboard{
		app:    app,
		status: status,
		info:   table,
	}
}

// Update refresh the dashboard with new telemetry.
func (d *Dashboard) Update(s *protocol.StatusResponse) {
	d.app.QueueUpdateDraw(func() {
		// Update Zone with Color
		zoneColor := "green"
		switch s.Zone {
		case protocol.SomaticZone_ZONE_YELLOW:
			zoneColor = "yellow"
		case protocol.SomaticZone_ZONE_RED:
			zoneColor = "red"
		}

		d.info.SetCell(1, 1, tview.NewTableCell(s.Zone.String()).SetTextColor(tcell.GetColor(zoneColor)))
		d.info.SetCell(2, 1, tview.NewTableCell(fmt.Sprintf("%d%%", s.PressureScore)))
		d.info.SetCell(3, 1, tview.NewTableCell(fmt.Sprintf("%d bytes", s.MemoryUsageBytes)))
		d.info.SetCell(4, 1, tview.NewTableCell(fmt.Sprintf("%d", s.HeapObjects)))
		d.info.SetCell(5, 1, tview.NewTableCell(fmt.Sprintf("%d", s.GoroutineCount)))
	})
}

// Run starts the dashboard application.
func (d *Dashboard) Run() error {
	d.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && event.Rune() == 'q' {
			d.app.Stop()
		}
		return event
	})

	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(d.status, 1, 1, false).
		AddItem(d.info, 0, 1, true).
		AddItem(tview.NewTextView().SetText("Press 'q' or Ctrl+C to exit"), 1, 1, false)

	return d.app.SetRoot(layout, true).Run()
}

// Stop closes the dashboard.
func (d *Dashboard) Stop() {
	d.app.Stop()
}
