package mule

import "github.com/nsf/termbox-go"

type view struct {
	mule *MULE

	// Current player location
	xpos int
	ypos int

	// Recent player positions
	xposq []int
	yposq []int
	iq    int
	iql   int

	delay int

	backing_rune []rune
	backing_fg   []termbox.Attribute
	backing_bg   []termbox.Attribute
	bounds       []bool
}

type continueType int

const (
	continueTypeStay = iota
	continueTypeSwitch
)

const (
	timeX0 int = 65
)

func (v *view) Print(x, y int, msg string, fg, bg termbox.Attribute, add bool, addb bool) {
	mg := v.mule
	for _, c := range msg {
		if add {
			i := y*mg.w + x
			v.backing_rune[i] = c
			v.backing_fg[i] = fg
			v.backing_bg[i] = bg
			if addb {
				v.bounds[i] = true
			}
		}
		termbox.SetCell(mg.x0+x, mg.y0+y, c, fg, bg)
		x++
	}
}

func (v *view) Banner(msg []string, fg, bg termbox.Attribute) {

	mg := v.mule
	y := 0
	for _, mv := range msg {

		for j := 0; j < 2*mg.w; j++ {
			termbox.SetCell(j, y, ' ', fg, bg)
		}

		x := 2
		for _, c := range mv {
			termbox.SetCell(x, y, c, fg, bg)
			x++
		}
		y++
	}
}

func (v *view) PrintTime(msg string) {
	for k, c := range msg {
		termbox.SetCell(timeX0+k, 2, c, termbox.ColorWhite, termbox.ColorBlack)
	}
	termbox.Flush()
}

func (v *view) DrawHline(x1, x2, y int, c rune, fg, bg termbox.Attribute) {
	mg := v.mule
	for x := x1; x <= x2; x++ {
		termbox.SetCell(mg.x0+x, mg.y0+y, c, fg, bg)
		i := y*mg.w + x
		v.bounds[i] = true
		v.backing_rune[i] = c
		v.backing_fg[i] = fg
		v.backing_bg[i] = bg
	}
}

func (v *view) DrawVline(x, y1, y2 int, c rune, fg, bg termbox.Attribute) {
	mg := v.mule
	for y := y1; y <= y2; y++ {
		termbox.SetCell(mg.x0+x, mg.y0+y, c, fg, bg)
		i := y*mg.w + x
		v.bounds[i] = true
		v.backing_rune[i] = c
		v.backing_fg[i] = fg
		v.backing_bg[i] = bg
	}
}

func (v *view) turn(p int, pr rune, prc termbox.Attribute, tf func(v *view, x, y int) location,
	kh func(*view, termbox.Event) continueType) location {

	mg := v.mule

	// Main event loop
	for cnt := 0; ; cnt++ {
	ax:
		select {
		case <-v.mule.mainTimer.C:
			return locTimeout

		case msg := <-v.mule.timerinfo:
			v.PrintTime(msg)

		case stat := <-v.mule.wumpusStatus:
			if v.mule.currentStage == stageLiveField {
				if stat.active {
					v.Print(stat.x, stat.y, "W", termbox.ColorWhite, termbox.ColorBlack, false, false)
				} else {
					v.RestorePoint(stat.x, stat.y)
				}
				termbox.Flush()
			}

		case ev := <-v.mule.eventQueue:
			if ev.Type == termbox.EventKey {
				newX := v.xpos
				newY := v.ypos
				switch {
				case ev.Ch == 'a':
					if mg.currentStage == stageLiveField {
						mg.assay_x = v.xpos
						mg.assay_y = v.ypos
						mg.Banner("Soil sample obtained, return to assay office for processing", 0)
					}
				case ev.Key == termbox.KeyArrowUp:
					if cnt%100 < 20*v.delay {
						continue
					}
					newY--
				case ev.Key == termbox.KeyArrowLeft:
					if cnt%100 < 20*v.delay {
						continue
					}
					newX--
				case ev.Key == termbox.KeyArrowRight:
					if cnt%100 < 20*v.delay {
						continue
					}
					newX++
				case ev.Key == termbox.KeyArrowDown:
					if cnt%100 < 20*v.delay {
						continue
					}
					newY++
				default:
					cont := kh(v, ev)
					if cont == continueTypeSwitch {
						return locStoreNone
					}
					if cont == continueTypeStay {
						break ax
					}
				}

				c := tf(v, newX, newY)

				// Check if we are entering a target
				if c == locStoreBlocked {
					break ax
				}
				if c != locStoreNone {
					return c
				}

				// Hit the edge of the screen
				if newX < 0 || newY < 0 || newX >= mg.w || newY >= mg.h {
					break ax
				}

				// Hit an obstacle
				if v.bounds[newY*mg.w+newX] {
					break ax
				}

				// Remove player piece at previous position
				for k := 0; k < v.iql; k++ {
					if v.xposq[k] != -1 && v.yposq[k] != -1 {
						i := v.yposq[k]*mg.w + v.xposq[k]
						termbox.SetCell(mg.x0+v.xposq[k], mg.y0+v.yposq[k], v.backing_rune[i],
							v.backing_fg[i], v.backing_bg[i])
					}
				}

				// Move the point
				v.iq = (v.iq + 1) % v.iql
				v.xposq[v.iq%v.iql] = newX
				v.yposq[v.iq%v.iql] = newY
				v.xpos = newX
				v.ypos = newY

				// Draw player piece at new position
				i := newY*mg.w + newX
				termbox.SetCell(mg.x0+newX, mg.y0+newY, pr, prc, v.backing_bg[i])

				// Draw the mule
				if v.mule.Model.Players[p].hasMule {
					v.DrawMule(p)
				}
				termbox.Flush()

			}
		default:
			// nothing
		}
	}

	panic("should not reach here")
	return -1
}

func (v *view) RestorePoint(x, y int) {
	mg := v.mule
	i := y*mg.w + x
	c := v.backing_rune[i]
	fg := v.backing_fg[i]
	bg := v.backing_bg[i]
	v.Print(x, y, string(c), fg, bg, false, false)
}

func (v *view) DrawPlayer(p, x, y int) {
	mg := v.mule
	i := y*mg.w + x
	v.Print(x, y, "Y", v.mule.PlayerColors[p], v.backing_bg[i], false, false)
}

func (v *view) RemoveMule(col termbox.Attribute) {
	mg := v.mule
	for k := 0; k < v.iql; k++ {
		if v.xposq[k] != -1 {
			i := v.yposq[k]*mg.w + v.xposq[k]
			v.Print(v.xposq[k], v.yposq[k], " ", v.backing_fg[i], v.backing_bg[i], false, false)

			// Draw back whatever was under the mule
			v.Print(v.xposq[k], v.yposq[k], string(v.backing_rune[i]), v.backing_fg[i], v.backing_bg[i], false, false)
		}
	}

	// Redraw the player
	i := v.ypos*mg.w + v.xpos
	v.Print(v.xpos, v.ypos, "Y", col, v.backing_bg[i], false, false)
}

func (v *view) DrawMule(p int) {

	mg := v.mule
	x0 := v.xposq[v.iq%v.iql]
	y0 := v.yposq[v.iq%v.iql]
	x1 := v.xposq[(v.iq+3-1)%v.iql]
	y1 := v.yposq[(v.iq+3-1)%v.iql]
	x2 := v.xposq[(v.iq+3-2)%v.iql]
	y2 := v.yposq[(v.iq+3-2)%v.iql]
	var c rune
	if y2 == y0 || y2 == -1 {
		c = '-'
	} else if x2 == x0 {
		c = '|'
	} else if (x2-x0)*(y2-y0) > 0 {
		c = '\\'
	} else {
		c = '/'
	}

	pl := v.mule.Model.Players[p]
	prc := v.mule.PlayerColors[p]

	if x1 != -1 {
		i := y1*mg.w + x1
		termbox.SetCell(mg.x0+x1, mg.y0+y1, c, prc, v.backing_bg[i])
	}
	if x2 != -1 {
		i := y2*mg.w + x2
		termbox.SetCell(mg.x0+x2, mg.y0+y2, pl.muleSymbol, prc, v.backing_bg[i])
	}
}

func (v *view) view_init() {
	mg := v.mule
	v.iql = 3
	m := mg.w * mg.h
	v.bounds = make([]bool, m)
	v.backing_rune = make([]rune, m)
	v.backing_fg = make([]termbox.Attribute, m)
	v.backing_bg = make([]termbox.Attribute, m)
	v.xposq = make([]int, v.iql)
	v.yposq = make([]int, v.iql)
}

func (v *view) Clear() {

	mg := v.mule
	bg := termbox.ColorBlack

	for i := 0; i < mg.h; i++ {
		for j := 0; j < mg.w; j++ {
			termbox.SetCell(mg.x0+j, mg.y0+i, ' ', bg, bg)
			k := mg.w*i + j
			v.backing_rune[k] = ' '
			v.backing_fg[k] = bg
			v.backing_bg[k] = bg
		}
	}
}
