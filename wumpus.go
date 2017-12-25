package mule

import (
	"math/rand"
	"time"
)

type wumpusInfo struct {
	x      int
	y      int
	active bool
}

func wumpusManager(mg *MULE) {

	fv := mg.Fieldview

	// Get the indices of the plots with mountains
	var xv, yv []int
	for i := 0; i < nrow; i++ {
		for j := 0; j < ncol; j++ {
			pl := mg.Model.GetPlot(i, j)
			if pl.Mountains > 0 {
				yv = append(yv, i)
				xv = append(xv, j)
			}
		}
	}

	for {
		// wait
		x := 2 + int(rand.Int63()%int64(10))
		time.Sleep(time.Duration(x) * time.Second)

		// random offset within the plot
		k := int(rand.Int63() % int64(len(xv)))
		i0 := int(rand.Int63() % int64(ploth-1))
		j0 := int(rand.Int63() % int64(plotw-1))

		fv.wumpusy = yv[k]*ploth + i0 + 1
		fv.wumpusx = xv[k]*plotw + j0 + 1
		fv.wumpusOut = true

		mg.wumpusStatus <- wumpusInfo{fv.wumpusx, fv.wumpusy, true}

		// wait
		x = 5 + int(rand.Int63()%int64(20))
		time.Sleep(time.Duration(x) * time.Second)
		fv.wumpusOut = false
		mg.wumpusStatus <- wumpusInfo{fv.wumpusx, fv.wumpusy, false}
	}
}
