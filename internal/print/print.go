package print

import (
	"fmt"

	"github.com/Simek13/satelliteApp/internal/satellites"
	"github.com/Simek13/satelliteApp/internal/sort"
)

func PrintSatelliteMeasurementTimes(sats map[string]satellites.Satellite) {
	for _, sat := range sats {
		fmt.Println(sat.GetSatellite().Id, "-", sat.MeasurementTime())
	}
}

func PrintSatelliteCalculationAverages(ionoAvg, NDVIAvg, radAvg, altAvg, salAvg map[string]float64) {
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
