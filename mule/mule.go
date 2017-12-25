package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/EmmaShedden/mule"
	"github.com/nsf/termbox-go"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	gameinfo := mule.GetGameInfo()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	eventQueue := make(chan termbox.Event)

	// Capture and pre-screen the events
	go func() {
		for {
			x := termbox.PollEvent()

			// Check for killing the game
			if x.Type == termbox.EventKey {
				if x.Key == termbox.KeyCtrlC {
					termbox.Close()
					os.Exit(0)
				}
			}

			eventQueue <- x
		}
	}()

	gameinfo.PlayerColors = []termbox.Attribute{termbox.ColorRed,
		termbox.ColorGreen, termbox.ColorYellow, termbox.ColorMagenta}

	mm := mule.NewModel(gameinfo)
	sv := mule.NewStoreView()
	fv := mule.NewFieldView()
	av := mule.NewAuctionView()
	mg := mule.NewMule(mm, sv, fv, av, eventQueue, gameinfo)

	fid, err := os.Create("mule.log")
	if err != nil {
		panic(err)
	}
	defer fid.Close()
	mg.Logger = log.New(fid, "", log.Lshortfile)

	mg.Play()
}
