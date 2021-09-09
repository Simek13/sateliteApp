package main

import (
	"fmt"
	"os"
	"satelites/csvreader"
	"satelites/math"
	"satelites/sort"
	"strconv"
	"time"
)

type Satelite interface {
	measurementTime() time.Duration
	compute() ([]float64, []float64, []float64, []float64)
	getSatelite() *BasicSatelite
}

type BasicSatelite struct {
	id               string
	timestamps       []time.Time
	ionoIndexes      []float64
	ndviIndexes      []float64
	radiationIndexes []float64
}

type EaSatelite struct {
	sat       *BasicSatelite
	altitudes []float64
}

type VcSatelite struct {
	sat         *BasicSatelite
	vegetations []string
}

type SsSatelite struct {
	sat           *BasicSatelite
	seaSalinities []float64
}

func (sat *BasicSatelite) measurementTime() time.Duration {
	minD := math.MinDate(sat.timestamps)
	maxD := math.MaxDate(sat.timestamps)
	timeDiff := maxD.Sub(minD)
	return timeDiff
}

func (sat *BasicSatelite) compute() ([]float64, []float64, []float64, []float64) {
	values := sat.ionoIndexes
	min := math.Min(values)
	max := math.Max(values)
	avg := math.Avg(values)
	ionoCalc := [3]float64{min, max, avg}
	values = sat.ndviIndexes
	min = math.Min(values)
	max = math.Max(values)
	avg = math.Avg(values)
	ndviCalc := [3]float64{min, max, avg}
	values = sat.radiationIndexes
	min = math.Min(values)
	max = math.Max(values)
	avg = math.Avg(values)
	radiationCalc := [3]float64{min, max, avg}

	return ionoCalc[:], ndviCalc[:], radiationCalc[:], nil

}

func (eaSat *EaSatelite) measurementTime() time.Duration {
	return eaSat.sat.measurementTime()
}

func (eaSat *EaSatelite) compute() ([]float64, []float64, []float64, []float64) {
	ionoCalc, ndviCalc, radiationCalc, _ := eaSat.sat.compute()
	values := eaSat.altitudes
	min := math.Min(values)
	max := math.Max(values)
	avg := math.Avg(values)
	altitudesCalc := [3]float64{min, max, avg}
	return ionoCalc, ndviCalc, radiationCalc, altitudesCalc[:]
}

func (eaSat *EaSatelite) getSatelite() *BasicSatelite {
	return eaSat.sat
}

func (vcSat *VcSatelite) measurementTime() time.Duration {
	return vcSat.sat.measurementTime()
}

func (vcSat *VcSatelite) compute() ([]float64, []float64, []float64, []float64) {
	ionoCalc, ndviCalc, radiationCalc, _ := vcSat.sat.compute()
	return ionoCalc, ndviCalc, radiationCalc, nil
}

func (vcSat *VcSatelite) getSatelite() *BasicSatelite {
	return vcSat.sat
}

func (ssSat *SsSatelite) measurementTime() time.Duration {
	return ssSat.sat.measurementTime()
}

func (ssSat *SsSatelite) compute() ([]float64, []float64, []float64, []float64) {
	ionoCalc, ndviCalc, radiationCalc, _ := ssSat.sat.compute()
	values := ssSat.seaSalinities
	min := math.Min(values)
	max := math.Max(values)
	avg := math.Avg(values)
	salinitiesCalc := [3]float64{min, max, avg}
	return ionoCalc, ndviCalc, radiationCalc, salinitiesCalc[:]
}

func (ssSat *SsSatelite) getSatelite() *BasicSatelite {
	return ssSat.sat
}

func main() {
	url := "https://raw.githubusercontent.com/sea43d/PythonEvaluation/master/satDataCSV2.csv"
	data, err := csvreader.ReadCsvFromUrl(url)
	if err != nil {
		panic(err)
	}

	satelites := make(map[string]Satelite)

	for _, row := range data[1:] {

		satId := row[0]
		if _, ok := satelites[satId]; !ok {
			sat := &BasicSatelite{id: satId,
				timestamps:       make([]time.Time, 0),
				ionoIndexes:      make([]float64, 0),
				ndviIndexes:      make([]float64, 0),
				radiationIndexes: make([]float64, 0),
			}
			switch satId {
			case "30J14":
				satelites[satId] = &EaSatelite{
					sat:       sat,
					altitudes: make([]float64, 0),
				}
			case "13A14", "6N14":
				satelites[satId] = &SsSatelite{
					sat:           sat,
					seaSalinities: make([]float64, 0),
				}
			default:
				satelites[satId] = &VcSatelite{
					sat:         sat,
					vegetations: make([]string, 0),
				}
			}
		}

		layout := "01-02-2006 15:04"
		tm, err := time.Parse(layout, row[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		satelite := satelites[satId]

		satelite.getSatelite().timestamps = append(satelite.getSatelite().timestamps, tm)

		if val, err := strconv.ParseFloat(row[2], 64); err == nil {
			satelite.getSatelite().ionoIndexes = append(satelite.getSatelite().ionoIndexes, val)
		}
		if val, err := strconv.ParseFloat(row[3], 64); err == nil {
			satelite.getSatelite().ndviIndexes = append(satelite.getSatelite().ndviIndexes, val)
		}
		if val, err := strconv.ParseFloat(row[4], 64); err == nil {
			satelite.getSatelite().radiationIndexes = append(satelite.getSatelite().radiationIndexes, val)
		}

		switch satelite.(type) {
		case *EaSatelite:
			if val, err := strconv.ParseFloat(row[5], 64); err == nil {
				satelite.(*EaSatelite).altitudes = append(satelite.(*EaSatelite).altitudes, val)
			}
		case *SsSatelite:
			if val, err := strconv.ParseFloat(row[5], 64); err == nil {
				satelite.(*SsSatelite).seaSalinities = append(satelite.(*SsSatelite).seaSalinities, val)
			}
		case *VcSatelite:
			satelite.(*VcSatelite).vegetations = append(satelite.(*VcSatelite).vegetations, row[5])
		}
	}

	// Date
	for _, sat := range satelites {
		fmt.Println(sat.getSatelite().id, "-", sat.measurementTime())
	}

	fmt.Println()

	ionoAvg := make(map[string]float64)
	NDVIAvg := make(map[string]float64)
	radAvg := make(map[string]float64)
	altAvg := make(map[string]float64)
	salAvg := make(map[string]float64)

	for id, sat := range satelites {

		fmt.Println("Satelite: ", id)
		ionoCalc, ndviCalc, radCalc, specCalc := sat.compute()
		fmt.Println("Ionosphere index:", ionoCalc[0], "(MIN)", ionoCalc[1], "(MAX)", ionoCalc[2], "(AVG)")
		ionoAvg[id] = ionoCalc[2]
		fmt.Println("NDVI index:", ndviCalc[0], "(MIN)", ndviCalc[1], "(MAX)", ndviCalc[2], "(AVG)")
		NDVIAvg[id] = ndviCalc[2]
		fmt.Println("Radiation index:", radCalc[0], "(MIN)", radCalc[1], "(MAX)", radCalc[2], "(AVG)")
		radAvg[id] = radCalc[2]
		switch sat.(type) {
		case *EaSatelite:
			fmt.Println("Earth altitude:", specCalc[0], "(MIN)", specCalc[1], "(MAX)", specCalc[2], "(AVG)")
			altAvg[id] = specCalc[2]
		case *SsSatelite:
			fmt.Println("Sea salinity index:", specCalc[0], "(MIN)", specCalc[1], "(MAX)", specCalc[2], "(AVG)")
			salAvg[id] = specCalc[2]
		}

		fmt.Println()

	}

	fmt.Println("@Ionosphere index:")
	for _, ss := range sort.Sort(ionoAvg) {
		fmt.Println(ss)
	}

	fmt.Println("@NDVI index:")
	for _, ss := range sort.Sort(NDVIAvg) {
		fmt.Println(ss)
	}

	fmt.Println("@Radiation index:")
	for _, ss := range sort.Sort(radAvg) {
		fmt.Println(ss)
	}

	fmt.Println("@Earth altitude:")
	for _, ss := range sort.Sort(altAvg) {
		fmt.Println(ss)
	}

	fmt.Println("@Sea salinity:")
	for _, ss := range sort.Sort(salAvg) {
		fmt.Println(ss)
	}

}
