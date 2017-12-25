package mule

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

// Colors
const (
	backgroundColor = termbox.ColorBlack
	boardColor      = termbox.ColorBlack

	// Offset for the MULE location within the plot
	starH = 4
	starV = 3

	animationSpeed = 100 * time.Millisecond

	homeSymbol     = "\u2302"
	mountainSymbol = "∧"
)

var (
	selectKeys = []rune{'a', 'd', 'g', 'j'}
)

type FieldView struct {
	view

	// Wumpus status
	wumpusOut bool
	wumpusx   int
	wumpusy   int

	// Random drift for the river
	rd []int

	// Mountain locations
	mt []int
}

func (fv *FieldView) Init() {

	fv.view_init()

	// River drift positions
	fv.rd = make([]int, nrow*ploth+1)
	for i := ploth*nrow/2 - 2*ploth/3; i >= 0; i-- {
		fv.rd[i] = fv.rd[i+1] + int(rand.Int63()%3-1)
		if fv.rd[i] > 3 {
			fv.rd[i] = 3
		} else if fv.rd[i] < -3 {
			fv.rd[i] = -3
		}
	}
	for i := ploth*(nrow/2+1) + 2*nrow/3; i < len(fv.rd); i++ {
		fv.rd[i] = fv.rd[i-1] + int(rand.Int63()%3-1)
		if fv.rd[i] > 3 {
			fv.rd[i] = 3
		} else if fv.rd[i] < -3 {
			fv.rd[i] = -3
		}
	}
}

func NewFieldView() *FieldView {
	fv := new(FieldView)
	return fv
}

func (fv *FieldView) DrawLandscape() {

	fg := termbox.ColorWhite
	bg := termbox.ColorBlue

	// The river
	xm := ncol * plotw / 2
	for y := 0; y < nrow*ploth; y++ {
		x := fv.rd[y] + xm - 2
		fv.Print(x, y, "~~~~", fg, bg, true, false)
	}

	// The store
	fv.FillPlot(nrow/2, ncol/2, 'O', termbox.ColorWhite, boardColor)

	// The mountains
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			lev := fv.mule.Model.GetPlot(i, j).Mountains
			if lev == 0 {
				continue
			}
			s := strings.Repeat(mountainSymbol, lev)
			x := j*plotw + 1
			y := i*ploth + 1
			fv.Print(x, y, s, termbox.ColorWhite|termbox.AttrBold, termbox.ColorBlack, true, false)
		}
	}
}

func (fv *FieldView) ShowProduction() {
	mg := fv.mule
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			plt := fv.mule.Model.GetPlot(i, j)
			if !plt.Owned {
				continue
			}
			pcol := fv.mule.PlayerColors[plt.Owner]

			x0 := j * plotw
			y0 := i * ploth
			qm := fmt.Sprintf("%d", plt.Production)
			x := x0 + 1
			y := y0 + ploth - 2
			ii := y*mg.w + x
			bg := fv.backing_bg[ii]
			fv.Print(x, y, qm, pcol|termbox.AttrReverse, bg, false, false)
		}
	}
}

func (fv *FieldView) PlayerTurn(p, r int, side location) (bool, location) {

	// Start at either the right or left side of the store
	fv.ypos = ploth * nrow / 2
	if side == locStoreLeft {
		fv.xpos = plotw*ncol/2 - plotw/2 - 1
		fv.xposq = []int{fv.xpos, -1, -1}
	} else if side == locStoreRight {
		fv.xpos = plotw*ncol/2 + plotw/2 + 1
		fv.xposq = []int{fv.xpos, -1, -1}
	} else {
		panic("Invalid store location code\n")
	}
	fv.yposq = []int{fv.ypos, fv.ypos, fv.ypos}
	fv.iq = 0

	fv.DrawPlayer(p, fv.xpos, fv.ypos)
	fv.DrawLandscape()
	fv.DrawOwnedPlots()
	termbox.Flush()

	// Reference to the player, needs to be a reference as we will
	// mutate it below
	py := fv.mule.Model.Players[p]

	// Current player color
	prc := fv.mule.PlayerColors[p]

	// Location handler
	lh := func(v *view, x, y int) location {

		i := v.ypos / ploth
		j := v.xpos / plotw

		// Update the speed delay parameter
		plt := v.mule.Model.GetPlot(i, j)
		v.delay = plt.Mountains
		if plt.River {
			v.delay = 2
		}

		// Check the wumpus
		if fv.wumpusOut && (v.ypos == fv.wumpusy) && (v.xpos == fv.wumpusx) {
			amt := 100 * (r + 4) / 4
			mg := fmt.Sprintf("You caught the wumpus and earned $%d!", amt)
			fv.wumpusOut = false
			fv.RestorePoint(v.xpos, v.ypos)
			py.money += amt
			fv.Banner([]string{mg}, termbox.ColorWhite, termbox.ColorBlack)
			termbox.Flush()
		}

		// Check if we are entering the store
		i0 := nrow / 2
		y0 := i0 * ploth
		y1 := y0 + ploth
		j0 := ncol / 2
		x0 := j0 * plotw
		x1 := x0 + plotw
		if y > y0 && y < y1-1 {
			if x == x0 {
				return locStoreLeft
			} else if x == x1-1 {
				return locStoreRight
			}
		}

		return locStoreNone
	}

	// Key handler. Spacebar installs/releases mule.
	kh := func(v *view, ev termbox.Event) continueType {

		if ev.Key != termbox.KeySpace {
			return continueTypeStay
		}
		if !py.hasMule {
			return continueTypeStay
		}

		// Position of Mule
		xm := fv.xposq[(fv.iq+1)%fv.iql]
		ym := fv.yposq[(fv.iq+1)%fv.iql]

		// Position of Mule within plot
		xr := xm % plotw
		yr := ym % ploth

		// In correct position to install Mule
		if xr == starH && yr == starV {

			// Plot we are trying to install the Mule on
			j := xm / plotw
			i := ym / ploth
			pl := fv.mule.Model.GetPlot(i, j)

			if pl.Owned && pl.Owner == p {
				pl.MuleStatus = py.muleOutfitType
				py.hasMule = false
				py.muleOutfitType = outfitNone
				fv.RemoveMule(prc)
				fv.DrawOwnedPlots() // adds the Mule symbol to the plot

				// Move the player next to the mule
				fv.RestorePoint(fv.xpos, fv.ypos)
				fv.ypos = i*ploth + starV
				fv.xpos = j*plotw + starH + 1
				fv.iq = 0
				fv.xposq = []int{fv.xpos, -1, -1}
				fv.yposq = []int{fv.ypos, -1, -1}
				fv.DrawPlayer(p, fv.xpos, fv.ypos)

				fv.Banner([]string{"MULE successfully installed"}, termbox.ColorWhite, boardColor)
				fv.DrawOwnedPlots()
				termbox.Flush()
				return continueTypeStay
			}
		}

		// Mule escapes
		py.hasMule = false
		py.muleOutfitType = outfitNone
		msg := []string{"Your MULE escaped!"}
		fv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
		fv.RemoveMule(prc)
		termbox.Flush()
		return continueTypeStay
	}

	// Play until we enter the store
	for {
		c := fv.turn(p, 'Y', prc, lh, kh)

		switch c {
		case locTimeout:
			py.hasMule = false
			py.muleOutfitType = outfitNone
			msg := []string{"You are out of time!"}
			fv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
			termbox.Flush()
			time.Sleep(3000 * time.Millisecond)
			return false, locStoreNone
		case locStoreLeft:
			return true, locStoreLeft
		case locStoreRight:
			return true, locStoreRight
		case locStoreNone:
			time.Sleep(1000 * time.Millisecond)
		default:
			panic("Invalid store location code in field\n")
		}
	}

	return false, locStoreNone
}

func (fv *FieldView) drawPlotIcon(plt *Plot) {
	mg := fv.mule
	x := plt.Col*plotw + starH
	y := plt.Row*ploth + starV
	ii := y*mg.w + x
	col := termbox.ColorWhite | termbox.AttrBold
	if plt.MuleStatus == outfitNone {
		fv.Print(x, y, homeSymbol, col, fv.backing_bg[ii], true, false)
	} else {
		r := osym[plt.MuleStatus]
		fv.Print(x, y, string(r), col, fv.backing_bg[ii], true, false)
	}
}

func (fv *FieldView) DrawOwnedPlots() {

	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {

			plt := fv.mule.Model.GetPlot(i, j)
			if !plt.Owned {
				continue
			}

			col := fv.mule.PlayerColors[plt.Owner]
			fv.HighlightPlot(i, j, 'X', col, true)
			fv.drawPlotIcon(plt)
		}
	}
}

func (fv *FieldView) FillPlot(i, j int, c rune, fg, bg termbox.Attribute) {

	// upper/left corner of plot
	x0 := j * plotw
	y0 := i * ploth

	for i := 0; i < ploth; i++ {
		for j := 0; j < plotw; j++ {
			x := x0 + j
			y := y0 + i
			fv.Print(x, y, string(c), fg, bg, true, true)
		}
	}
}

func (fv *FieldView) HighlightPlot(i, j int, c rune, fg termbox.Attribute, add bool) {

	mg := fv.mule

	// upper/left corner of plot
	x0 := j * plotw
	y0 := i * ploth

	// Top/bottom sides
	for d := 0; d < ploth; d += ploth - 1 {
		for k := 0; k < plotw; k++ {
			x := x0 + k
			y := y0 + d
			ii := y*mg.w + x
			bg := fv.backing_bg[ii]
			termbox.SetCell(mg.x0+x, mg.y0+y, c, fg, bg)
			if add {
				fv.backing_rune[ii] = c
				fv.backing_fg[ii] = fg
			}
		}
	}

	// Left/right sides
	for d := 0; d < plotw; d += plotw - 1 {
		for k := 0; k < ploth; k++ {
			x := x0 + d
			y := y0 + k
			ii := y*mg.w + x
			bg := fv.backing_bg[ii]
			termbox.SetCell(mg.x0+x, mg.y0+y, c, fg, bg)
			if add {
				fv.backing_rune[ii] = c
				fv.backing_fg[ii] = fg
			}
		}
	}
}

func (fv *FieldView) RestoreHighlightedPlot(i, j int) {

	mg := fv.mule

	// upper/left corner of plot
	x0 := j * plotw
	y0 := i * ploth

	// Top/bottom sides
	for d := 0; d < ploth; d += ploth - 1 {
		for k := 0; k < plotw; k++ {
			x := x0 + k
			y := y0 + d
			ii := y*mg.w + x
			c := fv.backing_rune[ii]
			fg := fv.backing_fg[ii]
			bg := fv.backing_bg[ii]
			termbox.SetCell(mg.x0+x, mg.y0+y, c, fg, bg)
		}
	}

	// Left/right sides
	for d := 0; d < plotw; d += plotw - 1 {
		for k := 0; k < ploth; k++ {
			x := x0 + d
			y := y0 + k
			ii := y*mg.w + x
			c := fv.backing_rune[ii]
			fg := fv.backing_fg[ii]
			bg := fv.backing_bg[ii]
			termbox.SetCell(mg.x0+x, mg.y0+y, c, fg, bg)
		}
	}
}

func (fv *FieldView) selectHit() (bool, int) {
	// Break the highlight time into 5 segments
	for k := 0; k < 5; k++ {
		select {
		case ev := <-fv.mule.eventQueue:
			if ev.Type == termbox.EventKey {
				for j := 0; j < fv.mule.nplayers; j++ {
					if ev.Ch == selectKeys[j] {
						return true, j
					}
				}
			}
		default:
			time.Sleep(animationSpeed)
		}
	}
	return false, -1
}

func (fv *FieldView) FlashPlot(i, j, n int) {

	plt := fv.mule.Model.GetPlot(i, j)
	col := termbox.ColorWhite
	if plt.Owned {
		col = fv.mule.PlayerColors[plt.Owner]
	}
	for k := 0; k < n; k++ {
		if k%2 == 0 {
			fv.HighlightPlot(i, j, 'X', col, false)
		} else {
			fv.HighlightPlot(i, j, 'X', termbox.ColorBlack, false)
		}
		termbox.Flush()
		time.Sleep(time.Second)
	}
}

func (fv *FieldView) FlashRow(row int) {

	for j := 0; j < ncol; j++ {
		fv.HighlightPlot(row, j, 'X', termbox.ColorCyan, false)
		termbox.Flush()
		time.Sleep(time.Second)
		fv.RestoreHighlightedPlot(row, j)
		termbox.Flush()
	}
}

func (fv *FieldView) SelectPlot(r int) {

	mg := fv.mule
	fv.DrawLandscape()
	fv.DrawOwnedPlots()

	fg := termbox.ColorWhite

	msg := make([]string, 2)
	msg[0] = fmt.Sprintf("Select plots for round %d, press space to start.\n", r+1)
	var w []string
	for j, na := range fv.mule.PlayerNames {
		w = append(w, fmt.Sprintf("%s %q", na, selectKeys[j]))
	}
	msg[1] = "Player keys: " + strings.Join(w, ", ")

	fv.Banner(msg, fg, boardColor)
	termbox.Flush()
	fv.mule.WaitForSpace()

	selected := make([]bool, 4)
	nSelected := 0
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			pl := fv.mule.Model.GetPlot(i, j)
			if nSelected == fv.mule.nplayers {
				return
			}

			// Skip over owned plots
			if pl.Owned {
				continue
			}

			// Skip over the store
			if i == nrow/2 && j == ncol/2 {
				continue
			}

			fv.DrawLandscape()
			fv.DrawOwnedPlots()
			fv.HighlightPlot(i, j, 'X', termbox.ColorWhite, false)
			termbox.Flush()

			hit, p := fv.selectHit()
			if hit && !selected[p] {
				pl.Owned = true
				pl.Owner = p
				selected[p] = true
				nSelected++
				time.Sleep(100 * time.Millisecond)
				mg.drainQueue()
			}
			fv.HighlightPlot(i, j, ' ', boardColor, false)
		}
	}
}
