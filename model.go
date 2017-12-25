package mule

import (
	"math/rand"
	"sort"
)

type Model struct {
	mule *MULE

	// The field plots, stored row-wise
	plots []*Plot

	Players []*Player

	// Store prices
	muleStorePrice     int
	foodStorePrice     int
	energyStorePrice   int
	smithoreStorePrice int
	crystiteStorePrice int

	// Goods in store
	storeFood     int
	storeEnergy   int
	storeSmithore int
	storeCrystite int
	storeMules    int
}

const (
	// MULE outfitting costs
	crystiteOutfitCost int = 100
	smithoreOutfitCost int = 75
	energyOutfitCost   int = 50
	foodOutfitCost     int = 25
)

type Player struct {
	model *Model

	pnum int

	money int

	// Quantities of goods
	Food     int
	Energy   int
	Smithore int
	Crystite int

	FoodDeficit   int
	EnergyDeficit int

	score int
	rank  int

	hasMule        bool
	muleOutfitType outfitType
	muleSymbol     rune

	requiredEnergy int
	requiredFood   int
	availableTime  int
}

type location int

const (
	locNone = iota
	locTimeout
	locStoreNone
	locStoreLeft
	locStoreRight
	locStoreAssay
	locStoreCrystite
	locStoreSmithore
	locStoreEnergy
	locStoreFood
	locStorePub
	locStoreMules
	locStoreBlocked
)

type buyResult int

const (
	buyResultNomules = iota
	buyResultNomoney
	buyResultReturned
	buyResultSuccess
)

type outfitResult int

const (
	outfitResultNomule = iota
	outfitResultAlreadyOutfitted
	outfitResultNomoney
	outfitResultSuccess
)

type outfitType int

const (
	outfitCrystite = iota
	outfitSmithore
	outfitEnergy
	outfitFood
	outfitNone
)

type pubResult int

const (
	pubResultNoMules = iota
	pubResultSuccess
)

var (
	otype_names = map[outfitType]string{outfitCrystite: "Crystite", outfitSmithore: "Smithore",
		outfitFood: "Food", outfitEnergy: "Energy"}
	oconv = map[location]outfitType{locStoreCrystite: outfitCrystite,
		locStoreSmithore: outfitSmithore, locStoreEnergy: outfitEnergy,
		locStoreFood: outfitFood}
	osym = map[outfitType]rune{outfitCrystite: 'C', outfitSmithore: 'S',
		outfitEnergy: 'E', outfitFood: 'F'}
)

type Plot struct {
	Owned      bool
	Owner      int
	MuleStatus outfitType
	Mountains  int
	Production int
	River      bool
	Crystite   int
	Row        int
	Col        int
}

func (py *Player) updateScore(mg *MULE) {

	score := py.money
	score += py.Food * mg.Model.foodStorePrice
	score += py.Energy * mg.Model.energyStorePrice
	score += py.Smithore * mg.Model.smithoreStorePrice
	score += py.Crystite * mg.Model.crystiteStorePrice

	for _, plt := range mg.Model.plots {
		if plt.Owned && plt.Owner == py.pnum {
			switch plt.MuleStatus {
			case outfitNone:
				score += 500
			case outfitFood:
				score += 525 + 35
			case outfitEnergy:
				score += 550 + 35
			case outfitSmithore:
				score += 575 + 35
			case outfitCrystite:
				score += 600 + 35
			}
		}
	}

	py.score = score
}

func (md *Model) updateStorePrices(r int) {

	md.updateRequiredFood(r)
	md.updateRequiredEnergy(false)

	reqFood := 0
	reqEnergy := 0
	totFood := md.storeFood
	totEnergy := md.storeEnergy
	for _, py := range md.Players {
		reqFood += py.requiredFood
		reqEnergy += py.requiredEnergy
		totFood += py.Food
		totEnergy += py.Energy
	}

	rat := float64(reqFood) / float64(totFood)
	if rat < 1 {
		rat = 1
	}
	md.foodStorePrice = int(30 * (0.25 + 0.75*rat))

	rat = float64(reqEnergy) / float64(totEnergy)
	if rat < 1 {
		rat = 1
	}
	md.energyStorePrice = int(25 * (0.25 + 0.75*rat))

	// smithore store price
	mules := md.storeMules + md.storeSmithore/2 // mule equivalents in store
	if mules >= 5 {
		md.smithoreStorePrice = 50
	} else {
		md.smithoreStorePrice = 50 + 100*(5-mules)
	}
	x := int(rand.Int63() % 100)
	switch {
	case x <= 6:
		md.smithoreStorePrice -= 14
	case x <= 6+24:
		md.smithoreStorePrice -= 7
	case x <= 6+24+40:
		// no change
	case x <= 2+24+40+24:
		md.smithoreStorePrice += 7
	default:
		md.smithoreStorePrice += 14
	}

	md.crystiteStorePrice = 50 + int(rand.Int63()%100)

	md.muleStorePrice = 2 * md.smithoreStorePrice
}

func (py *Player) getResourceByName(r string) int {

	switch r {
	case "food":
		return py.Food
	case "energy":
		return py.Energy
	case "smithore":
		return py.Smithore
	case "crystite":
		return py.Crystite
	default:
		panic("unknown resource type")
	}

	return 0 // can't reach here
}

func (md *Model) getStoreAmount(rtp resourceType) int {
	switch rtp {
	case food:
		return md.storeFood
	case energy:
		return md.storeEnergy
	case smithore:
		return md.storeSmithore
	case crystite:
		return md.storeCrystite
	default:
		md.mule.Logger.Printf("Unkown resource %v", rtp)
	}
	return -1 // Can't reach here
}

func (md *Model) getStoreSellPrice(rtp resourceType, r int) (int, int) {

	md.updateStorePrices(r)

	var amn, amx int
	switch rtp {
	case food:
		amn = md.foodStorePrice
		amx = amn + 35
	case energy:
		amn = md.energyStorePrice
		amx = amn + 35
	case smithore:
		amn = md.smithoreStorePrice
		amx = amn + 35
	case crystite:
		amn = md.crystiteStorePrice
		amx = amn + 140
	}

	return amn, amx
}

// addone=true adds one to the result, used in the auction stage so
// that the player buys an extra energy for a mule to be purchased on
// the next turn
func (md *Model) updateRequiredEnergy(addone bool) {
	for _, py := range md.Players {
		if addone {
			py.requiredEnergy = 1
		} else {
			py.requiredEnergy = 0
		}
	}
	for _, plt := range md.plots {
		if plt.Owned {
			if plt.MuleStatus == outfitFood || plt.MuleStatus == outfitSmithore || plt.MuleStatus == outfitCrystite {
				md.Players[plt.Owner].requiredEnergy++
			}
		}
	}
}

func (md *Model) updateRequiredFood(r int) {
	for _, py := range md.Players {
		py.requiredFood = 3 + r/4
	}
}

func (md *Model) MakeStoreMules() {
	for md.storeMules < 14 {
		if md.storeSmithore >= 2 {
			md.storeSmithore -= 2
			md.storeMules++
		} else {
			break
		}
	}
}

func (md *Model) DoProduction() {

	// Base production of goods
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			plt := md.GetPlot(i, j)
			plt.DoProduction()
		}
	}

	// Handle energy deficits
	md.updateRequiredEnergy(false)
	for p, py := range md.Players {
		ed := py.EnergyDeficit
		if ed > 0 {
			md.mule.Logger.Printf("Player %d had %d energy deficits", p, ed)
			var plv []*Plot
			for _, plt := range md.plots {
				if plt.Owned && plt.Owner == p {
					if plt.MuleStatus == outfitFood || plt.MuleStatus == outfitSmithore || plt.MuleStatus == outfitCrystite {
						plv = append(plv, plt)
					}
				}
			}

			for j := 0; j < ed; j++ {
				q := int(rand.Int63() % int64(len(plv)))
				plv[q].Production = 0
				copy(plv[q:], plv[q+1:])
				plv = plv[0 : len(plv)-1]
			}
		}
	}

	// Add production to player totals
	for _, plt := range md.plots {
		if plt.Owned {
			switch plt.MuleStatus {
			case outfitFood:
				md.Players[plt.Owner].Food += plt.Production
			case outfitEnergy:
				md.Players[plt.Owner].Energy += plt.Production
			case outfitSmithore:
				md.Players[plt.Owner].Smithore += plt.Production
			case outfitCrystite:
				md.Players[plt.Owner].Crystite += plt.Production
			}
		}
	}
}

func (plt *Plot) baseProduction() int {

	var base int
	switch plt.MuleStatus {
	case outfitFood:
		switch {
		case plt.Mountains > 0:
			base = 1
		case plt.River:
			base = 4
		default:
			base = 2
		}
	case outfitEnergy:
		switch {
		case plt.Mountains > 0:
			base = 1
		case plt.River:
			base = 2
		default:
			base = 3
		}
	case outfitSmithore:
		switch {
		case plt.Mountains > 0:
			base = 1 + plt.Mountains
		case plt.River:
			base = 0
		default:
			base = 1
		}
	case outfitCrystite:
		base = plt.Crystite
	}

	return base
}

func (plt *Plot) DoProduction() {

	if plt.Owned == false || plt.MuleStatus == outfitNone {
		plt.Production = 0
		return
	}

	base := plt.baseProduction()

	// Random term
	var y int

	x := int(rand.Int63() % 100)
	switch {
	case x < 6:
		y = -2
	case x < 6+24:
		y = -1
	case x < 6+24+40:
		y = 0
	case x < 6+24+40+24:
		y = 1
	default:
		y = 2
	}

	prod := base + y
	if prod < 0 {
		prod = 0
	}
	if prod > 8 {
		prod = 8
	}
	plt.Production = prod
}

func (py *Player) Outfit(otype outfitType) outfitResult {
	if !py.hasMule {
		return outfitResultNomule
	}
	if (py.muleOutfitType != outfitNone) && (py.muleOutfitType == otype) {
		return outfitResultAlreadyOutfitted
	}

	switch otype {
	case outfitFood:
		py.money -= foodOutfitCost
	case outfitEnergy:
		py.money -= energyOutfitCost
	case outfitSmithore:
		py.money -= smithoreOutfitCost
	case outfitCrystite:
		py.money -= crystiteOutfitCost
	}

	return outfitResultSuccess
}

func (p *Player) gamblePub(r int, mg *MULE) (pubResult, int) {
	if p.hasMule {
		return pubResultNoMules, 0
	}

	rb := 50 * (1 + r/4)
	amt := rb + int(rand.Int63()%int64(mg.timeRemaining))
	p.money += amt
	return pubResultSuccess, amt
}

func (p *Player) BuyMule() buyResult {
	if p.model.storeMules == 0 {
		return buyResultNomules
	}
	if p.hasMule {
		p.model.storeMules++
		p.money += p.model.muleStorePrice
		p.hasMule = false
		return buyResultReturned
	}
	if p.model.muleStorePrice > p.money {
		return buyResultNomoney
	}

	p.money -= p.model.muleStorePrice
	p.model.storeMules--
	p.hasMule = true
	p.muleSymbol = 'M'
	p.muleOutfitType = outfitNone
	return buyResultSuccess
}

func newDefaultPlayer(model *Model, p int) *Player {
	var py Player
	py.model = model
	py.Food = 4
	py.Energy = 2
	py.money = 1000
	py.pnum = p
	return &py
}

func (md *Model) setupPlots() {

	m := nrow * ncol
	md.plots = make([]*Plot, m)
	k := 0
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			md.plots[k] = new(Plot)
			md.plots[k].MuleStatus = outfitNone
			md.plots[k].Row = i
			md.plots[k].Col = j
			k++
		}
	}

	// River status
	for i := 0; i < nrow; i++ {
		md.GetPlot(i, ncol/2).River = true
	}

	// Add the mountain status
	for _, lev := range []int{1, 2, 3} {
		for ms := 0; ms < 4; {
			k := int(rand.Int63() % int64(m))

			// No mountains on the river or store
			if k%ncol == ncol/2 {
				continue
			}

			if md.plots[k].Mountains == 0 {
				ms++
				md.plots[k].Mountains = lev
			}
		}
	}

	// Add the Crystite deposits
	ix := selectFrom(4, nrow*ncol)
	for k := 0; k < 4; k++ {
		i := ix[k] / ncol
		j := ix[k] % ncol
		md.addCrystite(i, j, 3)
		md.addCrystite(i, j-1, 2)
		md.addCrystite(i, j+1, 2)
		md.addCrystite(i-1, j, 2)
		md.addCrystite(i+1, j, 2)
		md.addCrystite(i+1, j+1, 1)
		md.addCrystite(i+1, j-1, 1)
		md.addCrystite(i-1, j+1, 1)
		md.addCrystite(i-1, j-1, 1)
		md.addCrystite(i-2, j, 1)
		md.addCrystite(i+2, j, 1)
		md.addCrystite(i, j-2, 1)
		md.addCrystite(i, j+2, 1)
	}
}

func (md *Model) addCrystite(i, j, v int) {
	if i < 0 || i >= nrow {
		return
	}
	if j < 0 || j >= ncol {
		return
	}
	md.GetPlot(i, j).Crystite = v
}

// Randomly select k integers from m
func selectFrom(k, m int) []int {

	v := make([]int, m)
	for j := 0; j < m; j++ {
		v[j] = j
	}

	r := make([]int, k)
	for j := 0; j < k; {
		i := int(rand.Int63() % int64(m))
		if v[i] != -1 {
			r[j] = v[i]
			j++
			v[i] = -1
		}
	}

	return r
}

func NewModel(gi *GameInfo) *Model {
	md := new(Model)

	md.setupPlots()

	// Setup players
	md.Players = make([]*Player, len(gi.PlayerNames))
	for p := 0; p < len(gi.PlayerNames); p++ {
		md.Players[p] = newDefaultPlayer(md, p)
	}

	// Initial values
	md.storeMules = 14
	md.muleStorePrice = 100
	md.storeFood = 8
	md.storeEnergy = 8
	md.storeSmithore = 8
	md.storeCrystite = 0

	return md
}

func (md *Model) GetPlot(row, col int) *Plot {
	if (row > nrow) || (col > ncol) {
		panic("Invalid arguments to GetPlot")
	}
	return md.plots[row*ncol+col]
}

func (md *Model) DoConsumptionSpoilage(r int) {

	md.updateRequiredFood(r)
	md.updateRequiredEnergy(false)

	for p := 0; p < md.mule.nplayers; p++ {
		py := md.mule.Model.Players[p]

		// Food
		x := py.requiredFood
		if py.Food-x > 1 {
			x += (py.Food - x - 1) / 2 // spoilage
		}
		py.Food -= x
		if py.Food < 0 {
			py.FoodDeficit = -x
			py.Food = 0
		} else {
			py.FoodDeficit = 0
		}
		md.mule.Logger.Printf("Player %d lost %d food to consumption/spoilatge", p, x)

		// Energy
		x = py.requiredEnergy
		if py.Energy-x > 2 {
			x += (py.Energy - x - 2) / 2
		}
		py.Energy -= x
		if py.Energy < 0 {
			py.EnergyDeficit = -x
			py.Energy = 0
		} else {
			py.EnergyDeficit = 0
		}
		md.mule.Logger.Printf("Player %d lost %d energy to consumption/spoilage", p, x)
	}
}

func (md *Model) playerTurnTime(p, r int) int {

	// Food requirement
	fr := 3 + r/4

	py := md.Players[p]
	rat := float64(py.Food) / float64(fr)
	if rat > 1 {
		rat = 1
	}
	tm := 10 + int(91*rat)
	tm = int(float64(tm) * 0.6)

	md.mule.Logger.Printf("Player %d time: %d", p, tm)

	return tm
}

type rst struct {
	i int
	x float64
}

type rsl []*rst

func (a rsl) Len() int           { return len(a) }
func (a rsl) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a rsl) Less(i, j int) bool { return a[i].x < a[j].x }

func (md *Model) updatePlayerRanks() {

	n := md.mule.nplayers
	rk := make([]*rst, n)
	for p := 0; p < n; p++ {
		py := md.Players[p]
		rk[p] = &rst{p, float64(py.score) + 0.01*rand.NormFloat64()}
	}

	sort.Sort(rsl(rk))

	for p := 0; p < n; p++ {
		md.Players[p].rank = (n - rk[p].i) - 1
	}
}

// randomOwnedPlot returns a randomly-selected plot from among all
// plots that are owned by a player with rank at most maxrank.  If
// types is not nil, only plots of the specified types may be
// returned.
func (md *Model) randomOwnedPlot(maxrank int, types []outfitType) *Plot {
	players := md.Players
	var plv []*Plot
	for _, plt := range md.plots {
		if !plt.Owned {
			continue
		}
		if players[plt.Owner].rank > maxrank {
			continue
		}
		if types == nil {
			plv = append(plv, plt)
		} else {
			for _, t := range types {
				if plt.MuleStatus == t {
					plv = append(plv, plt)
				}
			}
		}
	}

	if len(plv) == 0 {
		return nil
	}

	k := int(rand.Int63() % int64(len(players)))
	return plv[k]
}
