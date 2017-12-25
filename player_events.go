package mule

// http://bringerp.free.fr/RE/Mule/mule_document.html#RandomPlayerEvent

import (
	"fmt"
	"math/rand"
)

func hasMule(p int, mg *MULE) bool {
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			pl := mg.Model.GetPlot(i, j)
			if pl.Owned && pl.Owner == p && pl.MuleStatus != outfitNone {
				return true
			}
		}
	}

	return false
}

func countMiningMules(p int, mg *MULE) int {
	m := 0
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			pl := mg.Model.GetPlot(i, j)
			if pl.Owned && pl.Owner == p && (pl.MuleStatus == outfitSmithore || pl.MuleStatus == outfitCrystite) {
				m++
			}
		}
	}

	return m
}

func countEnergyMules(p int, mg *MULE) int {
	m := 0
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			pl := mg.Model.GetPlot(i, j)
			if pl.Owned && pl.Owner == p && pl.MuleStatus == outfitEnergy {
				m++
			}
		}
	}

	return m
}

func loosePlot(p int, mg *MULE) bool {
	var plts []*Plot
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			pl := mg.Model.GetPlot(i, j)
			if pl.Owned && pl.Owner == p {
				plts = append(plts, pl)
			}
		}
	}

	if len(plts) == 0 {
		return false
	}

	k := int(rand.Int63() % int64(len(plts)))
	plts[k].Owned = false
	plts[k].MuleStatus = outfitNone

	return true
}

func freePlot(p int, mg *MULE) bool {
	var plts []*Plot
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			pl := mg.Model.GetPlot(i, j)
			if !pl.Owned {
				plts = append(plts, pl)
			}
		}
	}

	if len(plts) == 0 {
		return false
	}

	k := int(rand.Int63() % int64(len(plts)))
	plts[k].Owned = true
	plts[k].Owner = p
	return true
}

func countFood(p int, mg *MULE) int {
	q := 0
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			pl := mg.Model.GetPlot(i, j)
			if pl.Owned && pl.Owner == p && pl.MuleStatus == outfitFood {
				q++
			}
		}
	}

	return q
}

func draw(p, r int, mg *MULE) (string, bool) {

	var k int
	for {
		k = int(rand.Int63() % int64(22))
		if !mg.playerEventHappened[k] {
			mg.playerEventHappened[k] = true
			break
		}
	}

	// Base value
	m := 25 * (r/4 + 1)

	var msg string

	py := mg.Model.Players[p]

	switch k {
	case 0:
		msg = "YOU JUST RECEIVED A PACKAGE FROM YOUR HOME-WORLD RELATIVES CONTAINING 3 FOOD AND 2 ENERGY UNITS."
		py.Food += 3
		py.Energy += 2
	case 1:
		msg = "A WANDERING SPACE TRAVELER REPAID YOUR HOSPITALITY BY LEAVING TWO BARS OF SMITHORE."
		py.Smithore += 3
	case 2:
		if !hasMule(p, mg) {
			return "", false
		}
		msg = fmt.Sprintf("YOUR MULE WAS JUDGED \"BEST BUILT\" AT THE COLONY FAIR. YOU WON $%d.", 2*m)
		py.money += 2 * m
	case 3:
		if !hasMule(p, mg) {
			return "", false
		}
		msg = fmt.Sprintf("YOUR MULE WON THE COLONY TAP-DANCING CONTEST. YOU COLLECTED $%d.", 4*m)
		py.money += 4 * m
	case 4:
		y := countFood(p, mg)
		if y == 0 {
			return "", false
		}
		msg = fmt.Sprintf("YOUR MULE WON THE COLONY TAP-DANCING CONTEST. YOU COLLECTED $%d.", 2*m*y)
		py.money += 2 * m * y
	case 5:
		msg = fmt.Sprintf("THE COLONY AWARDED YOU $%d FOR STOPPING THE WART WORM INFESTATION.", 4*m)
		py.money += 4 * m
	case 6:
		msg = fmt.Sprintf("THE MUSEUM BOUGHT YOUR ANTIQUE PERSONAL COMPUTER FOR $%d.", 8*m)
		py.money += 8 * m
	case 7:
		msg = fmt.Sprintf("YOU WON THE COLONY SWAMP EEL EATING CONTEST AND COLLECTED $%d. (YUCK!)", 2*m)
		py.money += 2 * m
	case 8:
		msg = fmt.Sprintf("A CHARITY FROM YOUR HOME-WORLD TOOK PITY ON YOU AND SENT $%d.", 3*m)
		py.money += 3 * m
	case 9:
		msg = fmt.Sprintf("YOUR OFFWORLD INVESTMENTS IN ARTIFICIAL DUMBNESS PAID $%d IN DIVIDENDS.", 6*m)
		py.money += 6 * m
	case 10:
		msg = fmt.Sprintf("A DISTANT RELATIVE DIED AND LEFT YOU A VAST FORTUNE. BUT AFTER TAXES YOU ONLY GOT $%d.", 4*m)
		py.money += 4 * m
	case 11:
		msg = fmt.Sprintf("YOU FOUND A DEAD MOOSE RAT AND SOLD THE HIDE FOR $%d.", 2*m)
		py.money += 2 * m
	case 12:
		f := freePlot(p, mg)
		if !f {
			return "", false
		}
		msg = fmt.Sprintf("YOU RECEIVED AN EXTRA PLOT OF LAND TO ENCOURAGE COLONY DEVELOPMENT.")
	case 13:
		msg = "MISCHIEVOUS GLAC-ELVES BROKE INTO YOUR STORAGE SHED AND STOLE HALF YOUR FOOD."
		py.Food /= 2
	case 14:
		if !hasMule(p, mg) {
			return "", false
		}
		msg = fmt.Sprintf("ONE OF YOUR MULES LOST A BOLT. REPAIRS COST YOU $%d.", 3*m)
		py.money -= 3 * m
	case 15:
		y := countMiningMules(p, mg)
		if y > 0 {
			msg = fmt.Sprintf("YOUR MINING MULES HAVE DETERIORATED FROM HEAVY USE AND COST $%d EACH TO REPAIR. THE TOTAL COST IS $%d.",
				2*m, 2*m*y)
		} else {
			return "", false
		}
		py.money -= 2 * m * y
	case 16:
		y := countEnergyMules(p, mg)
		if y > 0 {
			msg = fmt.Sprintf("THE SOLAR COLLECTORS ON YOUR ENERGY MULES ARE DIRTY. CLEANING COST YOU $%d EACH FOR A TOTAL OF $%d.",
				m, m*y)
		} else {
			return "", false
		}
		py.money -= m * y
	case 17:
		msg = fmt.Sprintf("YOUR SPACE GYPSY INLAWS MADE A MESS OF THE TOWN. IT COST YOU $%d TO CLEAN IT UP.", 6*m)
		py.money -= 6 * m
	case 18:
		msg = fmt.Sprintf("FLYING CAT-BUGS ATE THE ROOF OFF YOUR HOUSE. REPAIRS COST $%d.", 4*m)
		py.money -= 4 * m
	case 19:
		msg = fmt.Sprintf("YOU LOST $%d BETTING ON THE TWO-LEGGED KAZINGA RACES.", 4*m)
		py.money += 4 * m
	case 20:
		msg = fmt.Sprintf("YOUR CHILD WAS BITTEN BY A BAT LIZARD AND THE HOSPITAL BILL COST YOU $%d.", 4*m)
		py.money += 4 * m
	case 21:
		if loosePlot(p, mg) {
			msg = "YOU LOST A PLOT OF LAND BECAUSE THE CLAIM WAS NOT RECORDED."
		} else {
			return "", false
		}
	}

	return msg, true
}

func (mg *MULE) genEvent(p, r int) string {

	if int(rand.Int63()%int64(100)) >= 28 {
		return ""
	}

	var msg string
	var f bool

	for {
		msg, f = draw(p, r, mg)
		if f {
			break
		}
	}

	return msg
}
