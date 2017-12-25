package mule

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

type roundEventType int

const (
	pestAttackEvent = iota
	pirateShipEvent
	acidRainEvent
	planetquakeEvent
	sunspotsEvent
	meteoriteEvent
	radiationEvent
	fireInStoreEvent
)

var (
	roundEventMax = []int{3, 2, 3, 3, 3, 2, 2, 2}
)

func (mg *MULE) doSunspots() (string, bool) {

	if mg.roundEventCounts[sunspotsEvent] >= roundEventMax[sunspotsEvent] {
		return "", false
	}
	mg.roundEventCounts[sunspotsEvent]++

	for _, plt := range mg.Model.plots {
		if plt.Owned && plt.MuleStatus == outfitEnergy {
			plt.Production += 3
			if plt.Production > 8 {
				plt.Production = 8
			}
		}
	}

	msg := "Sunspot activity! Energy production increased by 3 units."
	return msg, true
}

func (mg *MULE) doAcidRain() (string, bool) {

	if mg.roundEventCounts[acidRainEvent] >= roundEventMax[acidRainEvent] {
		return "", false
	}
	mg.roundEventCounts[acidRainEvent]++

	row := int(rand.Int63() % int64(nrow))

	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			if i == nrow/2 && j == ncol/2 {
				continue
			}
			plt := mg.Model.GetPlot(i, j)
			if !plt.Owned {
				continue
			}
			if i == row {
				if plt.MuleStatus == outfitFood {
					plt.Production += 4
				} else if plt.MuleStatus == outfitEnergy {
					plt.Production -= 2
				}
			} else {
				if plt.MuleStatus == outfitFood {
					plt.Production += 1
				} else if plt.MuleStatus == outfitEnergy {
					plt.Production -= 1
				}
			}
			if plt.Production > 8 {
				plt.Production = 8
			}
			if plt.Production < 0 {
				plt.Production = 0
			}
		}
	}

	msg := fmt.Sprintf("Acid rain storm!  Food output is up, energy reduced.")
	mg.Banner(msg, 0)
	mg.Fieldview.FlashRow(row)

	return "", true
}

func (mg *MULE) doPlanetquake() (string, bool) {

	if mg.roundEventCounts[planetquakeEvent] >= roundEventMax[planetquakeEvent] {
		return "", false
	}
	mg.roundEventCounts[planetquakeEvent]++

	m := nrow * ncol
	for {
		// Find a random mountain
		k := int(rand.Int63() % int64(m))
		if mg.Model.plots[k].Mountains == 0 {
			continue
		}
		j := k % ncol
		i := k / ncol

		// Move the mountains to a neighboring plot
		for {
			i1 := 2*int(rand.Int63()%2) - 1
			j1 := 2*int(rand.Int63()%2) - 1

			if i+i1 < 0 || i+i1 >= nrow {
				continue
			}
			if j+j1 < 0 || j+j1 >= ncol {
				continue
			}

			mg.Model.GetPlot(i+i1, j+j1).Mountains = mg.Model.GetPlot(i, j).Mountains
			mg.Model.GetPlot(i, j).Mountains = 0
			break
		}
		break
	}

	msg := "Planetquake! All mining output is reduced."
	return msg, true
}

func (mg *MULE) doPirateShip() (string, bool) {

	if mg.roundEventCounts[pirateShipEvent] >= roundEventMax[pirateShipEvent] {
		return "", false
	}
	mg.roundEventCounts[pirateShipEvent]++

	for p := 0; p < mg.nplayers; p++ {
		mg.Model.Players[p].Crystite = 0
	}

	msg := "Pirate ship! All crystite in the colony is lost!"
	return msg, true
}

func (mg *MULE) doFireInStore() (string, bool) {

	if mg.roundEventCounts[fireInStoreEvent] >= roundEventMax[fireInStoreEvent] {
		return "", false
	}
	mg.roundEventCounts[fireInStoreEvent]++

	mg.Model.storeFood = 0
	mg.Model.storeEnergy = 0
	mg.Model.storeSmithore = 0

	msg := "There was a fire in the store!"
	return msg, true
}

func (mg *MULE) doPestAttack() (string, bool) {

	players := mg.Model.Players
	var plv []*Plot
	for _, plt := range mg.Model.plots {
		if plt.Owned && players[plt.Owner].rank <= 1 && plt.MuleStatus == outfitFood {
			plv = append(plv, plt)
		}
	}

	if len(plv) == 0 {
		return "", false
	}

	k := int(rand.Int63() % int64(len(plv)))
	plt := plv[k]
	plt.Production = 0
	msg := fmt.Sprintf("Pest attack! %s lost all production from one food plot", mg.PlayerNames[plt.Owner])
	return msg, true
}

func (mg *MULE) doRadiation() (string, bool) {

	if mg.roundEventCounts[radiationEvent] >= roundEventMax[radiationEvent] {
		return "", false
	}
	mg.roundEventCounts[radiationEvent]++

	players := mg.Model.Players
	plt := mg.Model.randomOwnedPlot(1, nil)

	if plt == nil {
		return "", false
	}

	switch plt.MuleStatus {
	case outfitFood:
		players[plt.Owner].Food -= plt.Production
	case outfitEnergy:
		players[plt.Owner].Energy -= plt.Production
	case outfitSmithore:
		players[plt.Owner].Smithore -= plt.Production
	case outfitCrystite:
		players[plt.Owner].Crystite -= plt.Production
	}

	// Hold the message, then remove the plot
	msg := fmt.Sprintf("Radiation! %s lost a plot!", mg.PlayerNames[plt.Owner])
	time.Sleep(2 * time.Second)
	plt.Owned = false
	plt.MuleStatus = outfitNone
	mg.Fieldview.DrawOwnedPlots()
	termbox.Flush()
	time.Sleep(2 * time.Second)

	return msg, true
}

func (mg *MULE) doMeteorite() (string, bool) {

	for {
		i := int(rand.Int63() % int64(nrow))
		j := int(rand.Int63() % int64(ncol))
		if j == ncol/2 {
			continue
		}
		plt := mg.Model.GetPlot(i, j)
		plt.Production = 0
		plt.MuleStatus = outfitNone
		plt.Crystite = 4
		mg.Banner("A meteorite strike creates new crystite deposit!", 0)
		mg.Fieldview.FlashPlot(i, j, 5)
		x, y := mg.Fieldview.xpos, mg.Fieldview.ypos
		mg.Fieldview.RestorePoint(x, y)
		return "", true
	}

	return "", false // can't reach here
}

func (mg *MULE) DoRoundEvent(r int) {

	if r == 11 {
		mg.Banner("The ship has returned", 0)
		termbox.Flush()
		return
	}

	k := int(rand.Int63() % 20)
	var msg string
	var f bool
	for {
		switch {
		case k <= 3:
			msg, f = mg.doPestAttack()
		case k <= 3+2:
			msg, f = mg.doPirateShip()
		case k <= 3+2+3:
			msg, f = mg.doAcidRain()
		case k <= 3+2+3+3:
			msg, f = mg.doPlanetquake()
		case k <= 3+2+3+3+3:
			msg, f = mg.doSunspots()
		case k <= 3+2+3+3+3+2:
			msg, f = mg.doMeteorite()
		case k <= 3+2+3+3+3+2+2:
			msg, f = mg.doRadiation()
		default:
			msg, f = mg.doFireInStore()
		}
		if f {
			break
		}
	}

	mg.Fieldview.ShowProduction()
	termbox.Flush()
	if len(msg) > 0 {
		mg.Banner(msg, 0)
		mg.Banner("", 1)
		time.Sleep(5 * time.Second)
	}
	mg.Banner("Press space bar to continue", 1)
	mg.WaitForSpace()
}
