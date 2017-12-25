// http://bringerp.free.fr/RE/Mule/mule_document.html

package mule

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/nsf/termbox-go"
)

type stage int

const (
	stagePlotSelection = iota
	stageLiveStore
	stageLiveField
	stageAuction
)

const (
	// Rows and columns of plots in the field
	nrow = 5
	ncol = 9

	// width, height of a plot
	ploth = 6
	plotw = 9

	statusbar_y int = 2
)

type GameInfo struct {
	PlayerNames  []string
	PlayerColors []termbox.Attribute
}

type MULE struct {
	Model *Model

	Storeview   *StoreView
	Fieldview   *FieldView
	Auctionview *AuctionView

	// width/height of the main region
	w int
	h int

	// Upper left corner of main region
	x0 int
	y0 int

	PlayerNames  []string
	PlayerColors []termbox.Attribute
	nplayers     int

	wumpusStatus chan wumpusInfo

	hasAssay bool
	assay_x  int
	assay_y  int

	currentStage  stage
	round         int
	timeRemaining int

	eventQueue chan termbox.Event

	Logger *log.Logger

	// The keys are player event numbers that have already been selected
	playerEventHappened map[int]bool

	// Count the number of times each round event occured
	roundEventCounts []int

	timerinfo chan string
	mainTimer *time.Timer
	Timers    []*time.Timer
}

func NewMule(md *Model, sv *StoreView, fv *FieldView, av *AuctionView,
	q chan termbox.Event, gi *GameInfo) *MULE {

	mg := new(MULE)
	mg.Model = md
	mg.Storeview = sv
	mg.Fieldview = fv
	mg.Auctionview = av
	mg.eventQueue = q

	mg.PlayerNames = gi.PlayerNames
	mg.PlayerColors = gi.PlayerColors
	mg.nplayers = len(gi.PlayerNames)

	// upper left corner of play region
	mg.x0 = 2
	mg.y0 = 4

	// width, height of field
	mg.w = plotw * ncol
	mg.h = ploth*nrow + 4

	// Set this as the parent of these components
	md.mule = mg
	sv.mule = mg
	fv.mule = mg
	av.mule = mg

	sv.Init()
	fv.Init()

	mg.playerEventHappened = make(map[int]bool)
	mg.roundEventCounts = make([]int, 8)

	mg.timerinfo = make(chan string)

	mg.wumpusStatus = make(chan wumpusInfo)
	go wumpusManager(mg)

	return mg
}

func (mg *MULE) WaitForSpace() {
	time.Sleep(100 * time.Millisecond)
	mg.drainQueue()
	for {
		select {
		case ev := <-mg.eventQueue:
			if ev.Type == termbox.EventKey && ev.Key == termbox.KeySpace {
				return
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (mg *MULE) drainQueue() {
	for {
		select {
		case <-mg.eventQueue:
		default:
			return
		}
	}
}

func timeFunc(mg *MULE, tm int) func() {
	f := func() {
		mg.timeRemaining = tm
		var msg string
		msg = fmt.Sprintf("Time: %2ds", tm)
		mg.timerinfo <- msg
	}
	return f
}

func (mg *MULE) setupTimer(py *Player) {

	mg.mainTimer = time.NewTimer(time.Duration(py.availableTime) * time.Second)

	for j := 0; j <= py.availableTime; j += 1 {
		f := timeFunc(mg, py.availableTime-j)
		t := time.AfterFunc(time.Duration(j)*time.Second, f)
		mg.Timers = append(mg.Timers, t)
	}
}

func (mg *MULE) ClearTimers() {
	if mg.mainTimer != nil {
		mg.mainTimer.Stop()
	}
	for _, t := range mg.Timers {
		t.Stop()
	}
	mg.Timers = mg.Timers[:0]
}

func (mg *MULE) PlayerTurn(p, r int) {

	first := true
	var side location
	var cont bool

	py := mg.Model.Players[p]
	py.availableTime = mg.Model.playerTurnTime(p, r)
	mg.ClearTimers()
	mg.setupTimer(py)
	mg.Fieldview.PrintTime(fmt.Sprintf("Time: %2ds    ", py.availableTime))
	mg.hasAssay = false

	mg.Fieldview.Clear()
	mg.updateStatusBar(p)
	mg.Storeview.initLive(p, locStoreLeft)
	mg.Storeview.DrawStore()
	termbox.Flush()
	evx := mg.genEvent(p, r)
	if evx != "" {
		mg.Logger.Printf("Player %d: %s\n", p, evx)
		mg.Banner(mg.PlayerNames[p]+": "+evx, 0)
		mg.Banner("", 1)
		termbox.Flush()
		time.Sleep(3 * time.Second)
		mg.Banner("Press space to start", 1)
		termbox.Flush()
	} else {
		msg := fmt.Sprintf("%s -- press space to start",
			mg.PlayerNames[p])
		mg.Banner(msg, 0)
		mg.Banner("", 1)
		termbox.Flush()
	}
	mg.WaitForSpace()
	mg.Banner("", 0)
	mg.Banner("", 1)

	for {
		// In store
		mg.currentStage = stageLiveStore
		mg.Fieldview.Clear()
		mg.updateStatusBar(p)
		mg.Storeview.initLive(p, side)
		cont, side = mg.Storeview.PlayerTurn(p, r, first, side)
		if !cont {
			break
		}
		first = false

		// In field
		mg.currentStage = stageLiveField
		mg.Storeview.Clear()
		mg.updateStatusBar(p)
		cont, side = mg.Fieldview.PlayerTurn(p, r, side)
		if !cont {
			break
		}
	}
	mg.ClearTimers()
}

func (mg *MULE) PlotSelection(r int) {
	mg.currentStage = stagePlotSelection
	mg.Fieldview.Clear()
	mg.Fieldview.DrawLandscape()
	mg.Fieldview.DrawOwnedPlots()
	mg.clearStatusBar()
	mg.Fieldview.SelectPlot(r)
	mg.Fieldview.DrawOwnedPlots()
	termbox.Flush()
	time.Sleep(2000 * time.Millisecond)
}

func (mg *MULE) DoProduction(r int) {

	mg.Model.DoProduction()

	mg.Fieldview.Clear()
	mg.Fieldview.DrawLandscape()
	mg.Fieldview.DrawOwnedPlots()
	mg.Fieldview.ShowProduction()
	mg.clearStatusBar()

	msg := fmt.Sprintf("Round %d production, press space to continue", r+1)
	mg.Banner(msg, 0)
	termbox.Flush()
	mg.WaitForSpace()
}

func (mg *MULE) DoAuction(r int) {

	mg.Fieldview.Clear()

	mg.Auctionview.Init(crystite, r)
	f := mg.Auctionview.DoDeclaration(r)
	if f {
		mg.Auctionview.DoAuction(r)
	}

	mg.Auctionview.Init(smithore, r)
	f = mg.Auctionview.DoDeclaration(r)
	if f {
		mg.Auctionview.DoAuction(r)
	}

	mg.Auctionview.Init(food, r)
	f = mg.Auctionview.DoDeclaration(r)
	if f {
		mg.Auctionview.DoAuction(r)
	}

	mg.Auctionview.Init(energy, r)
	f = mg.Auctionview.DoDeclaration(r)
	if f {
		mg.Auctionview.DoAuction(r)
	}
}

func (mg *MULE) Play() {

	// Loop over rounds
	for r := 0; r < 12; r++ {

		mg.round = r
		mg.Logger.Printf("Starting round %d", r+1)

		mg.PlotSelection(r)

		for p := 0; p < len(mg.PlayerNames); p++ {
			mg.PlayerTurn(p, r)
		}

		mg.Model.DoConsumptionSpoilage(r)
		mg.DoProduction(r)
		mg.DoRoundEvent(r)
		if r < 11 {
			mg.DoAuction(r)
		}

		// Need to do this here; if players sell smithore to
		// the store during the auction, the mules will be
		// available on the next round of player turns
		mg.Model.MakeStoreMules()

		mg.DoLeaderboard()
	}
}

func (mg *MULE) Banner(msg string, y int) {
	for k, c := range msg {
		termbox.SetCell(k+2, y, c, termbox.ColorWhite, termbox.ColorBlack)
	}
	for k := len(msg); k < 100; k++ {
		termbox.SetCell(k+2, y, ' ', termbox.ColorBlack, termbox.ColorBlack)
	}
	termbox.Flush()
}

func (mg *MULE) Print(x, y int, msg string, fg, bg termbox.Attribute) {
	for k, c := range msg {
		termbox.SetCell(x+k, y, c, fg, bg)
	}
	termbox.Flush()
}

func (mg *MULE) PrintMain(x, y int, msg string, fg, bg termbox.Attribute) {
	for k, c := range msg {
		termbox.SetCell(mg.x0+x+k, mg.y0+y, c, fg, bg)
	}
	termbox.Flush()
}

func GetGameInfo() *GameInfo {

	var nplayers int

	for {
		fmt.Print("\nHow many players (2-4): ")
		var nps string
		fmt.Scanln(&nps)
		var err error
		nplayers, err = strconv.Atoi(nps)
		if err == nil && nplayers >= 2 && nplayers <= 4 {
			break
		}
		fmt.Printf("You must enter a number between 2 and 4\n")
	}

	pnms := make([]string, nplayers)

	for j := 0; j < nplayers; j++ {
		for {
			fmt.Printf("\nWhat is the name of player %d? ", j+1)
			fmt.Scanln(&pnms[j])
			if len(pnms[j]) > 0 {
				break
			}
		}
	}

	gi := new(GameInfo)
	gi.PlayerNames = pnms

	return gi
}

func (mg *MULE) updateStatusBar(p int) {

	py := mg.Model.Players[p]

	h := "Money: %5d Food: %2d Energy: %2d Smithore: %2d Crystite: %2d"
	s := fmt.Sprintf(h, py.money, py.Food, py.Energy, py.Smithore, py.Crystite)

	fg := termbox.ColorWhite
	bg := termbox.ColorBlack

	for k, c := range s {
		termbox.SetCell(2+k, statusbar_y, c, fg, bg)
	}
	termbox.Flush()
}

func (mg *MULE) clearStatusBar() {
	bg := termbox.ColorBlack
	for k := 0; k < 100; k++ {
		termbox.SetCell(k, statusbar_y, ' ', bg, bg)
	}
	termbox.Flush()
}
