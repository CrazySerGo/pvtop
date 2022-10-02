package prevotes

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"log"
	"os"
	"time"
)

func DrawScreen(network string, voteChan chan []VoteState, pctChan chan float64, SummaryChan chan string) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	pctGuage := widgets.NewGauge()

	p := widgets.NewParagraph()
	p.Title = network

	lists := make([]*widgets.List, 3)
	for i := range lists {
		lists[i] = widgets.NewList()
		lists[i].Border = false
	}
	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(0.1,
			ui.NewCol(1.0/2, p),
			ui.NewCol(1.0/2, pctGuage),
		),
		ui.NewRow(0.9,
			ui.NewCol(.9/3, lists[0]),
			ui.NewCol(.9/3, lists[1]),
			ui.NewCol(1.2/3, lists[2]),
		),
	)
	ui.Render(grid)

	refresh := false
	tick := time.NewTicker(100 * time.Millisecond)
	uiEvents := ui.PollEvents()

	for {
		select {

		case <-tick.C:
			if !refresh {
				continue
			}
			refresh = false
			ui.Render(grid)

		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				ui.Clear()
				ui.Close()
				os.Exit(0)
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)
			}

		case votes := <-voteChan:
			refresh = true
			split, max := splitVotes(votes)
			for i := 0; i < max; i++ {
				lists[i].Rows = make([]string, len(split[i]))
				for j, voter := range split[i] {
					missing := "❌"
					if voter.Voted {
						missing = "✅"
					}
					lists[i].Rows[j] = fmt.Sprintf("%-3s %s", missing, voter.Description)
				}
			}

		case pct := <-pctChan:
			refresh = true
			pctGuage.Percent = int(pct * 100)

		case summary := <-SummaryChan:
			refresh = true
			p.Text = summary

		}
	}
}

func splitVotes(votes []VoteState) ([][]VoteState, int) {
	split := make([][]VoteState, 0)
	var max int
	switch {
	case len(votes) < 50:
		max = 1
		split = append(split, votes)
	case len(votes) < 100:
		max = 2
		split = append(split, votes[:50])
		split = append(split, votes[50:])
	default:
		max = 3
		split = append(split, votes[:50])
		split = append(split, votes[50:100])
		split = append(split, votes[100:])
	}
	return split, max
}
