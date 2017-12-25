package mule

import (
	"fmt"
	"strings"

	"github.com/nsf/termbox-go"
)

func (mg *MULE) getShortageWarnings() []string {

	mg.Model.updateRequiredFood(mg.round + 1)
	mg.Model.updateRequiredEnergy(false)

	var energy, food int
	for _, py := range mg.Model.Players {
		energy += py.requiredEnergy
		food += py.requiredFood
	}

	var v []string
	if food > mg.Model.storeFood {
		v = append(v, "food")
	}
	if energy > mg.Model.storeEnergy {
		v = append(v, "energy")
	}

	var msgs []string
	if len(v) > 0 {
		m := fmt.Sprintf("The colony has a shortage of %s!", strings.Join(v, " and "))
		msgs = append(msgs, m)
	}

	if mg.Model.storeMules < 4 {
		m := "The store has a shortage of smithore for MULEs!"
		msgs = append(msgs, m)
	}

	return msgs
}

func (mg *MULE) DoLeaderboard() {

	fg := termbox.ColorWhite
	bg := termbox.ColorBlack

	for p := 0; p < mg.nplayers; p++ {
		mg.Model.Players[p].updateScore(mg)
		mg.Model.updatePlayerRanks()
	}

	// Clear the screen
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			termbox.SetCell(x, y, ' ', bg, bg)
		}
	}

	for q := 0; q < mg.nplayers; q++ {

		// Display the players in rank order
		var col termbox.Attribute
		var py *Player
		var p int
		for p = 0; p < mg.nplayers; p++ {
			py = mg.Model.Players[p]
			col = mg.PlayerColors[p]
			if py.rank == q {
				break
			}
		}

		m := 3 + 7*p
		mg.Print(2, m, mg.PlayerNames[p], col, bg)
		mg.Print(18, m, fmt.Sprintf("%5d", py.score), col, bg)
		mg.Print(2, m+1, "  Money", col, bg)
		mg.Print(18, m+1, fmt.Sprintf("%5d", py.money), col, bg)
		mg.Print(2, m+2, "  Food", col, bg)
		mg.Print(18, m+2, fmt.Sprintf("%5d", py.Food), col, bg)
		mg.Print(2, m+3, "  Energy", col, bg)
		mg.Print(18, m+3, fmt.Sprintf("%5d", py.Energy), col, bg)
		mg.Print(2, m+4, "  Smithore", col, bg)
		mg.Print(18, m+4, fmt.Sprintf("%5d", py.Smithore), col, bg)
		mg.Print(2, m+5, "  Crystite", col, bg)
		mg.Print(18, m+5, fmt.Sprintf("%5d", py.Crystite), col, bg)
	}

	msgs := mg.getShortageWarnings()
	for y, msg := range msgs {
		mg.Print(1, y, msg, fg, bg)
	}

	mg.Print(1, 32, "Press space to continue", fg, bg)

	mg.WaitForSpace()

	// Clear the screen
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			termbox.SetCell(x, y, ' ', bg, bg)
		}
	}
}
