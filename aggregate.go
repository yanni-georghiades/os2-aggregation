package main

import (
	"fmt"
	"math"
	"gonum.org/v1/gonum/stat/distuv"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
    "gonum.org/v1/plot/vg"
)

type ThreePointEstimate struct {
	lowPoint float64
	midPoint float64
	highPoint float64
}

type RiskEventInput struct {
	estimate ThreePointEstimate
	likelihood float64
}

type RiskItemInput struct {
	estimate ThreePointEstimate
}

type Distribution struct {
	points []float64
	minPoint float64
	maxPoint float64
}

func main() {
	// copied from the "labor" item in the excel spreadsheet
	testRiskItemInputs := []RiskItemInput {
		RiskItemInput {
			ThreePointEstimate {
				lowPoint: 7.5,
				midPoint: 10.0, 
				highPoint: 15.0,
			},
		},
		RiskItemInput {
			ThreePointEstimate {
				lowPoint: 6.0,
				midPoint: 8.5, 
				highPoint: 12.5,
			},
		},
		RiskItemInput {
			ThreePointEstimate {
				lowPoint: 5.0,
				midPoint: 7.5, 
				highPoint: 13.0,
			},
		},
		RiskItemInput {
			ThreePointEstimate {
				lowPoint: 7.0,
				midPoint: 11.0, 
				highPoint: 15.0,
			},
		},
	}

	dist := logisticPoolingRiskItems(testRiskItemInputs)

	fmt.Println(dist)

	plotDistribution(dist)

}


func computeMean(est ThreePointEstimate) float64 {
	return (est.lowPoint + 4*est.midPoint + est.highPoint)/6
}

func computeStdDev(est ThreePointEstimate) float64 {
	return (est.highPoint - est.lowPoint) / 6
}

func computeAlpha(est ThreePointEstimate) float64 {
	return 4 * (est.midPoint - est.lowPoint) / (est.highPoint - est.lowPoint) + 1
}

func computeBeta(est ThreePointEstimate) float64 {
	return 4 * (est.highPoint - est.midPoint) / (est.highPoint - est.lowPoint) + 1
}

func linearPoolingRiskItems(inputs []RiskItemInput) ThreePointEstimate {
	ests := []ThreePointEstimate{}

	for i := 0; i < len(inputs); i++ {
		ests = append(ests, inputs[i].estimate)
	}

	return linearPooling(ests)
}

func linearPoolingRiskEvents(inputs []RiskEventInput) (ThreePointEstimate, float64) {
	ests := []ThreePointEstimate{}
	sumLikelihood := float64(0)

	for i := 0; i < len(inputs); i++ {
		ests = append(ests, inputs[i].estimate)
		sumLikelihood += inputs[i].likelihood
	}

	dist := linearPooling(ests)

	return dist, sumLikelihood / float64(len(inputs))
}


func linearPooling(inputs []ThreePointEstimate) ThreePointEstimate {
	pooled := ThreePointEstimate{float64(0), float64(0), float64(0)}

	for i := 0; i < len(inputs); i++ {
		pooled.lowPoint += inputs[i].lowPoint
		pooled.midPoint += inputs[i].midPoint
		pooled.highPoint += inputs[i].highPoint
	}

	pooled.lowPoint /= float64(len(inputs))
	pooled.midPoint /= float64(len(inputs))
	pooled.highPoint /= float64(len(inputs))

	return pooled
}


func logisticPoolingRiskItems(inputs []RiskItemInput) ThreePointEstimate {
	ests := []ThreePointEstimate{}

	for i := 0; i < len(inputs); i++ {
		ests = append(ests, inputs[i].estimate)
	}

	return logisticPooling(ests)
}

func logisticPoolingRiskEvents(inputs []RiskEventInput) (ThreePointEstimate, float64) {
	ests := []ThreePointEstimate{}
	sumLikelihood := float64(0)

	for i := 0; i < len(inputs); i++ {
		ests = append(ests, inputs[i].estimate)
		sumLikelihood += inputs[i].likelihood
	}

	dist := logisticPooling(ests)

	return dist, sumLikelihood /float64(len(inputs))
}


func logisticPooling(inputs []ThreePointEstimate) ThreePointEstimate {

	pooled := ThreePointEstimate{1, 1, 1}

	for i := 0; i < len(inputs); i++ {
		pooled.lowPoint *= inputs[i].lowPoint
		pooled.midPoint *= inputs[i].midPoint
		pooled.highPoint *= inputs[i].highPoint
	}

	pooled.lowPoint = math.Pow(pooled.lowPoint, float64(1) / float64(len(inputs)))
	pooled.midPoint = math.Pow(pooled.midPoint, float64(1) / float64(len(inputs)))
	pooled.highPoint = math.Pow(pooled.highPoint, float64(1) / float64(len(inputs)))

	return pooled
}


// constructInputBases takes the set of three point estimates as an input and returns 
func constructPlotPoints(input ThreePointEstimate) Distribution {

	points := []float64{}

	alpha := computeAlpha(input)
	beta := computeBeta(input)
	dist := distuv.Beta {
		Alpha: alpha,
		Beta: beta,
	}
	
	minPoint, maxPoint := determineOutputRange(input)

	for j := 0.; j <= 1.; j += .001  {
		point := dist.Prob(j) 
		points = append(points, point)
	}

	return Distribution{points, minPoint, maxPoint}
}


// determineOutputRange is used to dynamically size the range of x values depending on data provided
// using a conservative estimate of 8 standard deviations below and above the mean 
func determineOutputRange(input ThreePointEstimate) (float64, float64) {
	minPoint := input.lowPoint
	maxPoint := input.highPoint

	return minPoint, maxPoint
}

// plotDistribution plots a single distribution "points" and uses minPoint and maxPoint to determine the x values
func plotDistribution(est ThreePointEstimate) {
	dist := constructPlotPoints(est)

	p := plot.New()

	p.Title.Text = "Test plot" 

	pts := make(plotter.XYs, len(dist.points))
	for i := range pts {
		pts[i].X = dist.minPoint + float64(i)*((dist.maxPoint - dist.minPoint) / float64(len(dist.points)))
		pts[i].Y = dist.points[i]
	}

	fmt.Println(pts)

	s, err := plotter.NewScatter(pts)

	if err != nil {
        panic(err)
    }

    p.Add(s)

    if err := p.Save(6*vg.Inch, 6*vg.Inch, "distribution.png"); err != nil {
		panic(err)
	}
}

