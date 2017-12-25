package mule

import (
	"fmt"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

type AuctionView struct {
	mule *MULE

	barposu int
	barposl int

	minPrice int
	maxprice int

	barw int
	colw int

	aucType resourceType

	buySell      []typeBuySell
	pos          []int
	newpos       []int
	pastCritical []bool
	canSell      []bool
}

var (
	rtnames = map[resourceType]string{food: "Food", energy: "Energy",
		smithore: "Smithore", crystite: "Crystite"}
)

type resourceType int

const (
	food = iota
	energy
	smithore
	crystite
)

type typeBuySell int

const (
	buyer = iota
	seller
)

const (
	barmin        int = 3
	barmax        int = 20
	transactDelay     = 15
	eventDelay        = 50 * time.Millisecond

	// Extra offset for the auction area
	ax0 int = 5
	ay0 int = -8 // negative because we count from the bottom
)

var (
	pkeys = []rune{'1', 'q', 'f', 'v', '8', 'i', ';', '/'}
)

func NewAuctionView() *AuctionView {
	av := new(AuctionView)
	av.barw = 60
	av.colw = 12
	return av
}

func (av *AuctionView) Init(atp resourceType, r int) {
	mg := av.mule
	av.aucType = atp
	av.buySell = make([]typeBuySell, mg.nplayers)
	av.canSell = make([]bool, mg.nplayers)
	av.pastCritical = make([]bool, mg.nplayers)
	av.pos = make([]int, mg.nplayers)
	for k := 0; k < mg.nplayers; k++ {
		av.pos[k] = barmin
	}
	av.newpos = make([]int, mg.nplayers)
	av.barposu = barmax
	av.barposl = barmin

	av.minPrice, av.maxprice = av.mule.Model.getStoreSellPrice(atp, r)
}

func (av *AuctionView) PrintTime(msg string) {

}

func (av *AuctionView) printPlayerAmounts() {
	mg := av.mule

	bg := termbox.ColorBlack
	for p := 0; p < av.mule.nplayers; p++ {
		py := av.mule.Model.Players[p]
		var amt, rqamt int
		switch av.aucType {
		case food:
			amt = py.Food
			rqamt = py.requiredFood
		case energy:
			amt = py.Energy
			rqamt = py.requiredEnergy
		case smithore:
			amt = py.Smithore
			rqamt = 0
		case crystite:
			amt = py.Crystite
			rqamt = 0
		}

		// Goods
		s := fmt.Sprintf("%5d", amt)
		col := av.mule.PlayerColors[p]
		av.Print((p+1)*av.colw-3, mg.h-barmin+3, s, col, bg)

		// Required
		s = fmt.Sprintf("%5d", rqamt)
		av.Print((p+1)*av.colw-3, mg.h-barmin+4, s, col, bg)

		// Money
		s = fmt.Sprintf("%5d", py.money)
		av.Print((p+1)*av.colw-3, mg.h-barmin+5, s, col, bg)
	}
}

func (av *AuctionView) printLimitPrices() {
	mg := av.mule

	s := fmt.Sprintf("%5d", av.minPrice)
	av.Print(av.barw+3, mg.h-barmin, s, termbox.ColorWhite, termbox.ColorBlack)
	s = fmt.Sprintf("%5d", av.maxprice)
	av.Print(av.barw+3, mg.h-barmax, s, termbox.ColorWhite, termbox.ColorBlack)
}

func (av *AuctionView) Render() {
	av.printLimitPrices()
	av.printStoreAmount()
	av.printPlayerAmounts()
	termbox.Flush()
}

// draw the upper/lower limits as --- or ====
func (av *AuctionView) drawLimitsSelect() {
	mg := av.mule

	for p := 0; p < av.mule.nplayers; p++ {
		if av.buySell[p] == seller {
			av.pos[p] = barmax
		} else {
			av.pos[p] = barmin
		}
	}
	s := strings.Repeat("-", av.barw)
	av.Print(0, mg.h-barmax, s, termbox.ColorWhite, termbox.ColorBlack)
	av.Print(0, mg.h-barmin, s, termbox.ColorWhite, termbox.ColorBlack)
}

func (av *AuctionView) drawLimitsAuction() {
	mg := av.mule
	s := strings.Repeat("-", av.barw)
	av.Print(0, mg.h-barmax, s, termbox.ColorWhite, termbox.ColorBlack)
	av.Print(0, mg.h-barmin, s, termbox.ColorWhite, termbox.ColorBlack)
}

func (av *AuctionView) keyMsg() string {
	msg := fmt.Sprintf("Player keys: %s (%s/%s)", av.mule.PlayerNames[0],
		string(pkeys[0]), string(pkeys[1]))
	for j := 1; j < av.mule.nplayers; j++ {
		msg += fmt.Sprintf("  %s (%s/%s)", av.mule.PlayerNames[j],
			string(pkeys[2*j]), string(pkeys[2*j+1]))
	}
	return msg
}

func (av *AuctionView) skipAuction() bool {

	var x int
	switch av.aucType {
	case food:
		x += av.mule.Model.storeFood
		for p := 0; p < av.mule.nplayers; p++ {
			x += av.mule.Model.Players[p].Food
		}
	case energy:
		x += av.mule.Model.storeEnergy
		for p := 0; p < av.mule.nplayers; p++ {
			x += av.mule.Model.Players[p].Energy
		}
	case smithore:
		x += av.mule.Model.storeSmithore
		for p := 0; p < av.mule.nplayers; p++ {
			x += av.mule.Model.Players[p].Smithore
		}
	case crystite:
		x += av.mule.Model.storeCrystite
		for p := 0; p < av.mule.nplayers; p++ {
			x += av.mule.Model.Players[p].Crystite
		}
	}

	return x == 0
}

func (av *AuctionView) SetInitialDeclarationStatus() {

	mg := av.mule
	md := mg.Model

	if av.aucType == food {
		md.updateRequiredFood(mg.round + 1)
		for p := 0; p < mg.nplayers; p++ {
			py := md.Players[p]
			av.canSell[p] = py.Food > 0
			if py.Food > py.requiredFood {
				av.buySell[p] = seller
			} else {
				av.buySell[p] = buyer
			}
		}
	} else if av.aucType == energy {
		md.updateRequiredEnergy(true)
		for p := 0; p < mg.nplayers; p++ {
			py := md.Players[p]
			av.canSell[p] = py.Energy > 0
			if py.Energy > py.requiredEnergy {
				av.buySell[p] = seller
			} else {
				av.buySell[p] = buyer
			}
		}
	} else if av.aucType == smithore {
		for p := 0; p < mg.nplayers; p++ {
			py := md.Players[p]
			av.canSell[p] = py.Smithore > 0
			if py.Smithore > 0 {
				av.buySell[p] = seller
			} else {
				av.buySell[p] = buyer
			}
		}
	} else if av.aucType == crystite {
		for p := 0; p < mg.nplayers; p++ {
			py := md.Players[p]
			av.canSell[p] = py.Crystite > 0
			if py.Crystite > 0 {
				av.buySell[p] = seller
			} else {
				av.buySell[p] = buyer
			}
		}
	}
}

func (av *AuctionView) DoDeclaration(r int) bool {

	if av.skipAuction() {
		return false
	}

	mg := av.mule
	av.SetInitialDeclarationStatus()

	av.Clear()
	av.drawLimitsSelect()
	av.drawLabels()
	av.Render()
	bg := termbox.ColorBlack

	for p := 0; p < mg.nplayers; p++ {
		col := mg.PlayerColors[p]
		if av.buySell[p] == buyer {
			av.Print((p+1)*av.colw-1, mg.h-barmin+2, "Buy", col, bg)
			av.pos[p] = barmin - 1
		} else {
			av.Print((p+1)*av.colw-1, mg.h-barmax-2, "Sell", col, bg)
			av.pos[p] = barmax + 1
		}
	}

	av.drawPlayers()
	msg := fmt.Sprintf("Declare to buy/sell in the %s auction, press space to begin", rtnames[av.aucType])
	mg.Banner(msg, 0)
	msg = av.keyMsg()
	mg.Banner(msg, 1)
	mg.WaitForSpace()
	msg = fmt.Sprintf("Declaring buy/sell in the %s auction... (press backspace to end)", rtnames[av.aucType])
	mg.Banner(msg, 0)

	timer := time.NewTimer(time.Duration(5) * time.Second)

	for {
		select {
		case <-timer.C:
			return true

		case msg := <-mg.timerinfo:
			av.PrintTime(msg)

		case ev := <-mg.eventQueue:
			if ev.Type != termbox.EventKey {
				break
			}
			switch {
			case ev.Ch == pkeys[0]:
				// Player 1 up
				if mg.nplayers >= 1 && av.canSell[0] {
					av.buySell[0] = seller
				}
			case ev.Ch == pkeys[1]:
				// Player 1 down
				if mg.nplayers >= 1 {
					av.buySell[0] = buyer
				}
			case ev.Ch == pkeys[2]:
				// Player 2 up
				if mg.nplayers >= 2 && av.canSell[1] {
					av.buySell[1] = seller
				}
			case ev.Ch == pkeys[3]:
				// Player 2 down
				if mg.nplayers >= 2 {
					av.buySell[1] = buyer
				}
			case ev.Ch == pkeys[4]:
				// Player 3 up
				if mg.nplayers >= 3 && av.canSell[2] {
					av.buySell[2] = seller
				}
			case ev.Ch == pkeys[5]:
				// Player 3 down
				if mg.nplayers >= 3 {
					av.buySell[2] = buyer
				}
			case ev.Ch == pkeys[6]:
				// Player 4 up
				if mg.nplayers >= 4 && av.canSell[3] {
					av.buySell[3] = seller
				}
			case ev.Ch == pkeys[7]:
				// Player 4 down
				if mg.nplayers >= 4 {
					av.buySell[3] = buyer
				}
			case ev.Key == termbox.KeyBackspace2:
				mg.Banner("Declaring ended early!", 0)
				termbox.Flush()
				time.Sleep(1 * time.Second)
				return true
			}
		}

		av.drawLimitsSelect()
		for p := 0; p < mg.nplayers; p++ {
			col := mg.PlayerColors[p]
			var y1, y2, y3, y4 int
			var txt string
			if av.buySell[p] == seller {
				txt = "Sell"
				y1 = barmax + 1 // draw
				y2 = barmin - 1 // erase
				y3 = barmax + 2 // draw
				y4 = barmin - 2 // erase
			} else {
				txt = "Buy "
				y1 = barmin - 1
				y2 = barmax + 1
				y3 = barmin - 2
				y4 = barmax + 2
			}
			av.Print((p+1)*av.colw, mg.h-y1, "Y", col, bg)
			av.Print((p+1)*av.colw, mg.h-y2, " ", col, bg)
			av.Print((p+1)*av.colw-1, mg.h-y3, txt, col, bg)
			av.Print((p+1)*av.colw-1, mg.h-y4, "    ", col, bg)
		}
		termbox.Flush()
	}

	return true
}

func (av *AuctionView) drawPlayers() {
	mg := av.mule
	bg := termbox.ColorBlack
	for p := 0; p < mg.nplayers; p++ {
		col := mg.PlayerColors[p]
		av.Print((p+1)*av.colw, mg.h-av.pos[p], "Y", col, bg)
	}
}

func (av *AuctionView) removePlayers() {
	mg := av.mule
	bg := termbox.ColorBlack
	for p := 0; p < mg.nplayers; p++ {
		av.Print((p+1)*av.colw, mg.h-av.pos[p], " ", bg, bg)
	}
}

func (av *AuctionView) drawLabels() {
	mg := av.mule
	fg := termbox.ColorWhite
	bg := termbox.ColorBlack
	av.Print(av.barw+2, mg.h-barmin+3, rtnames[av.aucType]+"     ", fg, bg)
	av.Print(av.barw+2, mg.h-barmin+4, "Required     ", fg, bg)
	av.Print(av.barw+2, mg.h-barmin+5, "Money     ", fg, bg)
}

func (av *AuctionView) checkSellers() bool {

	mg := av.mule
	md := mg.Model

	for k := 0; k < mg.nplayers; k++ {
		if av.buySell[k] == seller {
			return true
		}
	}

	var x int
	switch av.aucType {
	case food:
		x = md.storeFood
	case energy:
		x = md.storeEnergy
	case smithore:
		x = md.storeSmithore
	case crystite:
		x = md.storeCrystite
	}
	if x > 0 {
		return true
	}

	mg.Banner("No sellers, no auction.", 0)
	time.Sleep(2 * time.Second)

	return false
}

func (av *AuctionView) DoAuction(r int) {

	mg := av.mule
	if !av.checkSellers() {
		return
	}

	// Set the position based on the results of the selection round
	for k := 0; k < mg.nplayers; k++ {
		if av.buySell[k] == seller {
			av.pos[k] = barmax + 1
		} else {
			av.pos[k] = barmin - 1
		}
	}

	av.barposu = barmax
	av.barposl = barmin

	av.Clear()
	av.Render()
	av.drawPlayers()

	msg := fmt.Sprintf("Press space to start the %s auction", rtnames[av.aucType])
	mg.Banner(msg, 0)
	msg = av.keyMsg()
	mg.Banner(msg, 1)
	av.drawLimitsAuction()
	av.drawPlayers()
	av.drawLabels()
	termbox.Flush()
	mg.WaitForSpace()
	mg.Banner(fmt.Sprintf("%s auction... (press backspace to end)", rtnames[av.aucType]), 0)

	av.RunAuction()
}

func (av *AuctionView) posMoney(pos int) int {
	rat := float64(pos-barmin) / float64(barmax-barmin)
	return int(av.minPrice + int(rat*float64(av.maxprice-av.minPrice)))
}

func (av *AuctionView) clipPos(pos []int) {

	mg := av.mule
	md := mg.Model

	for p := 0; p < mg.nplayers; p++ {
		if pos[p] < barmin-1 {
			pos[p] = barmin - 1
		}
		if pos[p] > barmax+1 {
			pos[p] = barmax + 1
		}
		py := md.Players[p]

		// Buyer can't afford
		if av.buySell[p] == buyer {
			for av.posMoney(pos[p]) > py.money {
				pos[p]--
				if pos[p] < barmin-1 {
					pos[p] = barmin - 1
					break
				}
			}
		}

		// Seller out of goods
		if av.buySell[p] == seller {
			switch av.aucType {
			case food:
				if py.Food == 0 {
					pos[p] = barmax + 1
				}
			case energy:
				if py.Energy == 0 {
					pos[p] = barmax + 1
				}
			case smithore:
				if py.Smithore == 0 {
					pos[p] = barmax + 1
				}
			case crystite:
				if py.Crystite == 0 {
					pos[p] = barmax + 1
				}
			}
		}
	}
}

func (av *AuctionView) Clear() {
	mg := av.mule
	bg := termbox.ColorBlack
	for i := 0; i < 40; i++ {
		for j := 0; j < 80; j++ {
			termbox.SetCell(j, mg.y0+i, ' ', bg, bg)
		}
	}
}

func (av *AuctionView) removeBars() {
	mg := av.mule
	fg := termbox.ColorWhite
	bg := termbox.ColorBlack
	m := av.barw
	if av.barposu != barmax && av.barposu != barmin {
		m += 10
	}
	s := strings.Repeat(" ", m)
	av.Print(0, mg.h-av.barposu, s, fg, bg)
	if av.barposl != av.barposu && av.barposl != barmin {
		if av.barposl != barmin {
			m += 10
		}
		s = strings.Repeat(" ", m)
		av.Print(0, mg.h-av.barposl, s, fg, bg)
	}
}

func (av *AuctionView) redrawBars(barposl, barposu int) {
	mg := av.mule
	fg := termbox.ColorWhite
	bg := termbox.ColorBlack

	av.removePlayers()
	av.removeBars()

	var s string
	if barposu == barposl {
		s = strings.Repeat("=", av.barw)
		av.Print(0, mg.h-barposu, s, fg, bg)
	} else {
		s = strings.Repeat("-", av.barw)
		av.Print(0, mg.h-barposl, s, fg, bg)
		av.Print(0, mg.h-barposu, s, fg, bg)
		s = fmt.Sprintf("%5d", av.posMoney(barposl))
		av.Print(av.barw+3, mg.h-barposl, s, fg, bg)
	}
	s = fmt.Sprintf("%5d", av.posMoney(barposu))
	av.Print(av.barw+3, mg.h-barposu, s, fg, bg)

	// Redraw the limit bars which may have been erased or
	// obscured
	if barposl != barmin {
		s = strings.Repeat("-", av.barw)
		av.Print(0, mg.h-barmin, s, fg, bg)
	}
	if barposl != barmax {
		s = strings.Repeat("-", av.barw)
		av.Print(0, mg.h-barmax, s, fg, bg)
	}

	av.barposu = barposu
	av.barposl = barposl
}

func (av *AuctionView) printStoreAmount() {

	mg := av.mule
	md := mg.Model

	var x int
	switch av.aucType {
	case food:
		x = av.mule.Model.storeFood
	case energy:
		x = md.storeEnergy
	case smithore:
		x = md.storeSmithore
	case crystite:
		x = md.storeCrystite
	}

	s := fmt.Sprintf("[%d]  ", x)
	av.Print(av.barw+10, mg.h-barmax, s, termbox.ColorWhite, termbox.ColorBlack)
}

func (av *AuctionView) AnySellers() bool {

	mg := av.mule
	for p := 0; p < mg.nplayers; p++ {
		if av.buySell[p] == seller {
			return true
		}
	}
	return false
}

func (av *AuctionView) RunAuction() {

	mg := av.mule
	md := mg.Model

	timer := time.NewTimer(time.Duration(30) * time.Second)
	copy(av.newpos, av.pos)
	haveSellers := av.AnySellers()

	// If at/past critical at the beinning, the player is not
	// required to move down twice
	md.updateRequiredFood(mg.round + 1)
	md.updateRequiredEnergy(true)
	for p := 0; p < mg.nplayers; p++ {
		py := md.Players[p]
		if av.aucType == food {
			if py.Food <= py.requiredFood {
				av.pastCritical[p] = true
			}
		} else if av.aucType == energy {
			if py.Energy <= py.requiredEnergy {
				av.pastCritical[p] = true
			}
		}
	}

	sellToStoren := 0
	sellToPlayern := 0
	buyFromStoren := 0

	// Main event loop
	for cnt := 0; ; cnt++ {

		select {
		case <-timer.C:
			mg.Banner("The auction is over!", 0)
			time.Sleep(2000 * time.Millisecond)
			return

		case msg := <-mg.timerinfo:
			av.PrintTime(msg)

		case ev := <-mg.eventQueue:
			if ev.Type != termbox.EventKey {
				break
			}
			switch {
			case ev.Ch == pkeys[0]:
				// Player 1 up
				if mg.nplayers >= 1 {
					av.newpos[0] = av.pos[0] + 1
				}
			case ev.Ch == pkeys[1]:
				// Player 1 down
				if mg.nplayers >= 1 {
					av.newpos[0] = av.pos[0] - 1
				}
			case ev.Ch == pkeys[2]:
				// Player 2 up
				if mg.nplayers >= 2 {
					av.newpos[1] = av.pos[1] + 1
				}
			case ev.Ch == pkeys[3]:
				// Player 2 down
				if mg.nplayers >= 2 {
					av.newpos[1] = av.pos[1] - 1
				}
			case ev.Ch == pkeys[4]:
				// Player 3 up
				if mg.nplayers >= 3 {
					av.newpos[2] = av.pos[2] + 1
				}
			case ev.Ch == pkeys[5]:
				// Player 3 down
				if mg.nplayers >= 3 {
					av.newpos[2] = av.pos[2] - 1
				}
			case ev.Ch == pkeys[6]:
				// Player 4 up
				if mg.nplayers >= 4 {
					av.newpos[3] = av.pos[3] + 1
				}
			case ev.Ch == pkeys[7]:
				// Player 4 down
				if mg.nplayers >= 4 {
					av.newpos[3] = av.pos[3] - 1
				}
			case ev.Key == termbox.KeyBackspace2:
				mg.Banner("Auction ended early!", 0)
				termbox.Flush()
				time.Sleep(1 * time.Second)
				return
			}
		}

		av.clipPos(av.newpos)

		// Find the new bar positions
		barposu := barmax
		barposl := barmin
		for k := 0; k < mg.nplayers; k++ {
			if av.buySell[k] == seller {
				if av.newpos[k] < barposu {
					barposu = av.newpos[k]
				}
			} else if av.buySell[k] == buyer {
				if av.newpos[k] > barposl {
					barposl = av.newpos[k]
				}
			}
		}
		if barposl > barposu {
			barposl = av.barposl
			barposu = av.barposu
		}

		// Move players back to their bar if they have passed it
		for k := 0; k < mg.nplayers; k++ {
			if av.buySell[k] == seller {
				if av.newpos[k] < barposu {
					av.newpos[k] = barposu
				}
			}
			if av.buySell[k] == buyer {
				if av.newpos[k] > barposl {
					av.newpos[k] = barposl
				}
			}
		}

		av.redrawBars(barposl, barposu)
		copy(av.pos, av.newpos)
		av.drawPlayers()
		storeAmt := md.getStoreAmount(av.aucType)

		if av.barposu == barmin {
			sellToStoren++
			if sellToStoren == transactDelay {
				av.sellToStore()
				sellToStoren = 0
				av.printPlayerAmounts()
				av.printStoreAmount()
			}
		} else if av.barposl == barmax && storeAmt == 0 && haveSellers {
			av.maxprice++
			av.printLimitPrices()
		} else if av.barposl == barmax {
			buyFromStoren++
			if buyFromStoren == transactDelay {
				av.buyFromStore()
				buyFromStoren = 0
				av.printPlayerAmounts()
				av.printStoreAmount()
			}
		} else if av.barposu == av.barposl && av.barposu < barmax && av.barposl > barmin {
			sellToPlayern++
			if sellToPlayern == transactDelay {
				av.sellToPlayer()
				sellToPlayern = 0
				av.printPlayerAmounts()
			}
		}

		termbox.Flush()
		time.Sleep(eventDelay)
	}
}

func (av *AuctionView) Print(x, y int, msg string, fg, bg termbox.Attribute) {
	mg := av.mule
	mg.PrintMain(ax0+x, ay0+y, msg, fg, bg)
}

func (av *AuctionView) sellToStore() {

	mg := av.mule
	md := mg.Model

	// TODO need to randomize order here
	for k := 0; k < mg.nplayers; k++ {
		if av.pos[k] == barmin && av.buySell[k] == seller {
			py := md.Players[k]
			switch av.aucType {
			case food:
				if py.Food <= py.requiredFood && !av.pastCritical[k] {
					av.pos[k] = barmax
					av.pastCritical[k] = true
				} else if py.Food > 0 {
					py.Food--
					py.money += av.minPrice
					md.storeFood++
				}
			case energy:
				if py.Energy <= py.requiredEnergy && !av.pastCritical[k] {
					av.pos[k] = barmax
					av.pastCritical[k] = true
				} else if py.Energy > 0 {
					py.Energy--
					py.money += av.minPrice
					md.storeEnergy++
				}
			case smithore:
				if py.Smithore > 0 {
					py.Smithore--
					py.money += av.minPrice
					md.storeSmithore++
				}
			case crystite:
				if py.Crystite > 0 {
					py.Crystite--
					py.money += av.minPrice
					md.storeCrystite++
				}
			}
		}
	}
}

func (av *AuctionView) buyFromStore() {

	mg := av.mule
	md := mg.Model

	// TODO need to randomize order here
	for k := 0; k < mg.nplayers; k++ {
		if av.pos[k] == barmax && av.buySell[k] == buyer {
			py := md.Players[k]
			switch av.aucType {
			case food:
				if md.storeFood > 0 {
					py.Food++
					md.storeFood--
					py.money -= av.maxprice
				}
			case energy:
				if md.storeEnergy > 0 {
					py.Energy++
					md.storeEnergy--
					py.money -= av.maxprice
				}
			case smithore:
				if md.storeSmithore > 0 {
					py.Smithore++
					md.storeSmithore--
					py.money -= av.maxprice
				}
			case crystite:
				if md.storeCrystite > 0 {
					py.Crystite++
					md.storeCrystite--
					py.money -= av.maxprice
				}
			}
		}
	}
}

func (av *AuctionView) sellToPlayer() {

	mg := av.mule
	md := mg.Model

	// Find a buyer/seller pair
	// TODO should randomize order
	var buyerp, sellerp int
	for k := 0; k < mg.nplayers; k++ {
		if av.buySell[k] == seller && av.pos[k] == av.barposl {
			sellerp = k
		}
	}
	for k := 0; k < mg.nplayers; k++ {
		if av.buySell[k] == buyer && av.pos[k] == av.barposu {
			buyerp = k
		}
	}

	pyb := md.Players[buyerp]
	pys := md.Players[sellerp]
	amt := av.posMoney(av.barposu)
	switch av.aucType {
	case food:
		if pys.Food <= pys.requiredFood && !av.pastCritical[sellerp] {
			av.pos[sellerp] = barmax
			av.pastCritical[sellerp] = true
			mg.drainQueue()
		} else if pys.Food > 0 {
			pyb.Food++
			pys.Food--
			pyb.money -= amt
			pys.money += amt
		}
	case energy:
		if pys.Energy <= pys.requiredEnergy && !av.pastCritical[sellerp] {
			av.pos[sellerp] = barmax
			av.pastCritical[sellerp] = true
			mg.drainQueue()
		} else if pys.Energy > 0 {
			pyb.Energy++
			pys.Energy--
			pyb.money -= amt
			pys.money += amt
		}
	case smithore:
		if pys.Smithore > 0 {
			pyb.Smithore++
			pys.Smithore--
			pyb.money -= amt
			pys.money += amt
		}
	case crystite:
		if pys.Crystite > 0 {
			pyb.Crystite++
			pys.Crystite--
			pyb.money -= amt
			pys.money += amt
		}
	}
}
