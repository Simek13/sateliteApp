package satelites

import (
	"time"

	"github.com/Simek13/satelliteApp/internal/math"
)

type Satellite interface {
	MeasurementTime() time.Duration
	Compute() ([]float64, []float64, []float64, []float64)
	GetSatellite() *BasicSatellite
}

type SatType int

const (
	Ea SatType = iota
	Vc
	Ss
)

type BasicSatellite struct {
	Id               string
	Timestamps       []time.Time
	IonoIndexes      []float64
	NdviIndexes      []float64
	RadiationIndexes []float64
	Duration         time.Duration
	IonoCalc         []float64
	NdviCalc         []float64
	RadiationCalc    []float64
	SatelliteType    SatType
}

type EaSatellite struct {
	BasicSatellite
	Altitudes     []float64
	AltitudesCalc []float64
}

type VcSatellite struct {
	BasicSatellite
	Vegetations []string
}

type SsSatellite struct {
	BasicSatellite
	SeaSalinities  []float64
	SalinitiesCalc []float64
}

func (sat *BasicSatellite) MeasurementTime() time.Duration {
	minD := math.MinDate(sat.Timestamps)
	maxD := math.MaxDate(sat.Timestamps)
	timeDiff := maxD.Sub(minD)
	sat.Duration = timeDiff
	return timeDiff
}

func (sat *BasicSatellite) Compute() ([]float64, []float64, []float64, []float64) {
	values := sat.IonoIndexes
	min := math.Min(values)
	max := math.Max(values)
	avg := math.Avg(values)
	sat.IonoCalc = []float64{min, max, avg}
	values = sat.NdviIndexes
	min = math.Min(values)
	max = math.Max(values)
	avg = math.Avg(values)
	sat.NdviCalc = []float64{min, max, avg}
	values = sat.RadiationIndexes
	min = math.Min(values)
	max = math.Max(values)
	avg = math.Avg(values)
	sat.RadiationCalc = []float64{min, max, avg}

	return sat.IonoCalc[:], sat.NdviCalc[:], sat.RadiationCalc[:], nil

}

func (sat *BasicSatellite) GetSatellite() *BasicSatellite {
	return sat
}

func (eaSat *EaSatellite) Compute() ([]float64, []float64, []float64, []float64) {
	ionoCalc, ndviCalc, radiationCalc, _ := eaSat.BasicSatellite.Compute()
	values := eaSat.Altitudes
	min := math.Min(values)
	max := math.Max(values)
	avg := math.Avg(values)
	eaSat.AltitudesCalc = []float64{min, max, avg}
	return ionoCalc, ndviCalc, radiationCalc, eaSat.AltitudesCalc[:]
}

func (ssSat *SsSatellite) Compute() ([]float64, []float64, []float64, []float64) {
	ionoCalc, ndviCalc, radiationCalc, _ := ssSat.BasicSatellite.Compute()
	values := ssSat.SeaSalinities
	min := math.Min(values)
	max := math.Max(values)
	avg := math.Avg(values)
	ssSat.SalinitiesCalc = []float64{min, max, avg}
	return ionoCalc, ndviCalc, radiationCalc, ssSat.SalinitiesCalc[:]
}
