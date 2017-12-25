package mule

import (
	"fmt"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	assay_y    int = 6
	crystite_y int = 8
	smithore_y int = 10
	energy_y   int = 12
	food_y     int = 14
	pub_y      int = 16
	mules_y    int = 18
	start_y    int = 21
)

const (
	storeWidth int = 55
	slotStop   int = 41
	timerPos   int = 40

	// Extra horizontal/vertical offsets for store
	sx0 int = 10
	sy0 int = 5
)

type StoreView struct {
	view

	outOfSlot bool
}

func NewStoreView() *StoreView {
	sv := new(StoreView)
	return sv
}

func (sv *StoreView) Init() {
	sv.view_init()
}

func (sv *StoreView) DrawStore() {

	fg := termbox.ColorWhite
	bg := termbox.ColorBlack

	sv.DrawHline(sx0, sx0+storeWidth, sy0, 'X', fg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-10, sy0+2, '-', fg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-10, sy0+4, '-', fg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-10, sy0+6, '-', fg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-10, sy0+8, '-', fg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-10, sy0+10, '-', fg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-10, sy0+12, '-', fg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-10, sy0+14, 'X', fg, bg)
	sv.DrawHline(sx0, sx0+storeWidth, sy0+18, 'X', fg, bg)

	// Keep people from going down the slots too far
	sv.DrawHline(sx0, sx0+storeWidth-16, assay_y, ' ', bg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-16, crystite_y, ' ', bg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-16, smithore_y, ' ', bg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-16, energy_y, ' ', bg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-16, food_y, ' ', bg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-16, pub_y, ' ', bg, bg)
	sv.DrawHline(sx0, sx0+storeWidth-16, mules_y, ' ', bg, bg)

	// Left/right store walls
	sv.DrawVline(sx0, sy0, sy0+14, 'X', fg, bg)
	sv.DrawVline(sx0+storeWidth, sy0, sy0+14, 'X', fg, bg)

	sv.Print(sx0+4, assay_y, "Assay office", fg, bg, true, false)

	msg := fmt.Sprintf("Outfit MULE for crystite ($%d)", crystiteOutfitCost)
	sv.Print(sx0+4, crystite_y, msg, fg, bg, true, false)

	msg = fmt.Sprintf("Outfit MULE for smithore ($%d)", smithoreOutfitCost)
	sv.Print(sx0+4, smithore_y, msg, fg, bg, true, false)

	msg = fmt.Sprintf("Outfit MULE for energy ($%d)", energyOutfitCost)
	sv.Print(sx0+4, energy_y, msg, fg, bg, true, false)

	msg = fmt.Sprintf("Outfit MULE for food ($%d)", foodOutfitCost)
	sv.Print(sx0+4, food_y, msg, fg, bg, true, false)

	msg = "Pub"
	sv.Print(sx0+4, pub_y, msg, fg, bg, true, false)

	msg = fmt.Sprintf("%d MULEs ($%d each)", sv.mule.Model.storeMules, sv.mule.Model.muleStorePrice)
	sv.Print(sx0+4, mules_y, msg, fg, bg, true, false)
}

func (sv *StoreView) initLive(p int, side location) {

	bg := termbox.ColorBlack

	// Start at either the right or left side of the store,
	// default to left side for first move.
	if side == locStoreLeft || side == locStoreNone || side == locNone {
		sv.xpos = sx0
		sv.ypos = start_y
	} else if side == locStoreRight {
		sv.xpos = sx0 + storeWidth
		sv.ypos = start_y
	} else {
		panic("Invalid store location")
	}
	sv.Print(sv.xpos, sv.ypos, "Y", sv.mule.PlayerColors[p], bg, false, false)
	sv.xposq = []int{sv.xpos, -1, -1}
	sv.yposq = []int{sv.ypos, -1, -1}
	sv.iq = 0

	sv.DrawStore()
	termbox.Flush()
}

// Too big, needs refactoring
func (sv *StoreView) PlayerTurn(p, r int, first bool, side location) (bool, location) {

	mg := sv.mule
	py := mg.Model.Players[p]

	// Location handler
	lh := func(v *view, x, y int) location {

		// Don't let the player leave the store with a mule
		if y > 17 && y < mg.h && ((x <= sx0) || (x >= storeWidth+sx0)) {
			if py.hasMule && py.muleOutfitType == outfitNone {
				msg := []string{"Can't leave store with a MULE that is not outfitted"}
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
				termbox.Flush()
				return locStoreBlocked
			}
		}

		if x == slotStop+sx0 {
			// Entering one of the store areas
			if sv.outOfSlot {
				// Prevent rapid cycling in/out of slot
				sv.outOfSlot = false
				switch y {
				case assay_y:
					return locStoreAssay
				case crystite_y:
					return locStoreCrystite
				case smithore_y:
					return locStoreSmithore
				case energy_y:
					return locStoreEnergy
				case food_y:
					return locStoreFood
				case pub_y:
					return locStorePub
				case mules_y:
					return locStoreMules
				}
			}
		} else if x == slotStop+sx0+5 {
			sv.outOfSlot = true
		}

		// Exiting the store
		if y > 17 && y < mg.h {

			// Leaving on the left side
			if x < sx0 {
				return locStoreLeft
			}

			// Leaving on the right side
			if x > sx0+storeWidth {
				return locStoreRight
			}
		}

		return locStoreNone
	}

	kh := func(v *view, ev termbox.Event) continueType {
		return continueTypeStay
	}

	prc := mg.PlayerColors[p]

	// Eventloop
	for {
		loc := sv.turn(p, 'Y', prc, lh, kh)

		switch {
		case loc == locTimeout:
			mg := "You are out of time!"
			if py.hasMule {
				mg += "  You lost your MULE!"
			}
			msg := []string{mg}
			py.hasMule = false
			py.muleOutfitType = outfitNone
			sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
			termbox.Flush()
			time.Sleep(3000 * time.Millisecond)
			return false, locStoreNone

		case loc == locStoreLeft:
			return true, locStoreLeft

		case loc == locStoreRight:
			return true, locStoreRight

		// Maybe buy a mule
		case loc == locStoreMules:
			b := py.BuyMule()
			switch b {
			case buyResultNomules:
				msg := []string{"The store has no mules"}
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
				termbox.Flush()
			case buyResultNomoney:
				msg := []string{"You do not have enuogh money to buy a mule"}
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
				termbox.Flush()
			case buyResultReturned:
				msg := []string{"You returned your MULE to the store"}
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
				sv.RemoveMule(prc)
				sv.xposq = []int{sv.xpos, -1, -1}
				sv.yposq = []int{sv.ypos, -1, -1}
				sv.iq = 0
				sv.DrawStore()        // update mule count
				mg.updateStatusBar(p) // update money
				termbox.Flush()
			case buyResultSuccess:
				msg := []string{"You bought a mule"}
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
				sv.xposq = []int{sv.xpos, sv.xpos - 2, sv.xpos - 1}
				sv.yposq = []int{mules_y, mules_y, mules_y}
				sv.iq = 0
				py.muleSymbol = 'M'
				py.muleOutfitType = outfitNone
				sv.DrawMule(p)
				sv.DrawStore()        // update mule count
				mg.updateStatusBar(p) // update money
				termbox.Flush()
			default:
				panic("Invalid store location code in store\n")
			}
		case loc == locStoreAssay:
			if mg.hasAssay {
				i := mg.assay_y / ploth
				j := mg.assay_x / plotw
				plt := mg.Model.GetPlot(i, j)
				var msg string
				switch plt.Crystite {
				case 0:
					msg = "There is no crystite"
				case 1:
					msg = "Crystite level is low"
				case 2:
					msg = "Crystite level is medium"
				case 3:
					msg = "Crystite level is high"
				case 4:
					msg = "Crystite level is very high"
				}
				mg.Banner(msg, 0)
				mg.hasAssay = false
			} else {
				mg.hasAssay = true
				mg.Banner("Visit a plot and press 'a' to obtain soil sample", 0)
			}
		case loc == locStoreCrystite || loc == locStoreSmithore || loc == locStoreEnergy || loc == locStoreFood:
			// All outfit shops
			b := py.Outfit(oconv[loc])
			otp := oconv[loc]         // the type of outfitting
			oname := otype_names[otp] // the name of the type of outfitting
			osy := osym[otp]          // the MULE symbol
			switch b {
			case outfitResultNomule:
				msg := []string{"You don't have a MULE"}
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
				termbox.Flush()
			case outfitResultNomoney:
				msg := []string{fmt.Sprintf("You don't have enough money to outfit a MULE for %s", oname)}
				termbox.Flush()
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
			case outfitResultAlreadyOutfitted:
				msg := []string{fmt.Sprintf("Your MULE is already outfitted for %s", oname)}
				termbox.Flush()
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
			case outfitResultSuccess:
				msg := []string{fmt.Sprintf("Your MULE has been outfitted for %s", oname)}
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
				py.muleSymbol = osy
				py.muleOutfitType = otp
				sv.DrawMule(p)
				mg.updateStatusBar(p)
				termbox.Flush()
			default:
				panic("Invalid code in outfit\n")
			}
		case loc == locStorePub:
			b, amt := py.gamblePub(r, mg)
			switch {
			case b == pubResultNoMules:
				msg := []string{fmt.Sprintf("No MULEs allowed in the pub")}
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
				termbox.Flush()
			case b == pubResultSuccess:
				mg.ClearTimers()
				msg := []string{fmt.Sprintf("You won $%d gambling!", amt)}
				mg.updateStatusBar(p)
				sv.Print(sv.xpos, sv.ypos, "\u263A", mg.PlayerColors[p], termbox.ColorBlack,
					false, false)
				sv.Banner(msg, termbox.ColorWhite, termbox.ColorBlack)
				termbox.Flush()
				time.Sleep(4000 * time.Millisecond)
				return false, locStoreNone
			default:
				panic("Invalid code in pub")
			}
		}
	}

	return false, locStoreNone
}
