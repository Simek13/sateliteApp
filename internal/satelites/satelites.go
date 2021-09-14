package satelites

import (
	"time"

	"github.com/Simek13/sateliteApp/internal/math"
)

type Satelite interface {
	MeasurementTime() time.Duration
	Compute() ([]float64, []float64, []float64, []float64)
	GetSatelite() *BasicSatelite
}

type BasicSatelite struct {
	Id               string
	Timestamps       []time.Time
	IonoIndexes      []float64
	NdviIndexes      []float64
	RadiationIndexes []float64
	Duration         time.Duration
	IonoCalc         []float64
	NdviCalc         []float64
	RadiationCalc    []float64
}

type EaSatelite struct {
	Sat           *BasicSatelite
	Altitudes     []float64
	AltitudesCalc []float64
}

type VcSatelite struct {
	Sat         *BasicSatelite
	Vegetations []string
}

type SsSatelite struct {
	Sat            *BasicSatelite
	SeaSalinities  []float64
	SalinitiesCalc []float64
}

func (sat *BasicSatelite) MeasurementTime() time.Duration {
	minD := math.MinDate(sat.Timestamps)
	maxD := math.MaxDate(sat.Timestamps)
	timeDiff := maxD.Sub(minD)
	sat.Duration = timeDiff
	return timeDiff
}

func (sat *BasicSatelite) Compute() ([]float64, []float64, []float64, []float64) {
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

func (eaSat *EaSatelite) MeasurementTime() time.Duration {
	return eaSat.Sat.MeasurementTime()
}

func (eaSat *EaSatelite) Compute() ([]float64, []float64, []float64, []float64) {
	ionoCalc, ndviCalc, radiationCalc, _ := eaSat.Sat.Compute()
	values := eaSat.Altitudes
	min := math.Min(values)
	max := math.Max(values)
	avg := math.Avg(values)
	eaSat.AltitudesCalc = []float64{min, max, avg}
	return ionoCalc, ndviCalc, radiationCalc, eaSat.AltitudesCalc[:]
}

func (eaSat *EaSatelite) GetSatelite() *BasicSatelite {
	return eaSat.Sat
}

func (vcSat *VcSatelite) MeasurementTime() time.Duration {
	return vcSat.Sat.MeasurementTime()
}

func (vcSat *VcSatelite) Compute() ([]float64, []float64, []float64, []float64) {
	ionoCalc, ndviCalc, radiationCalc, _ := vcSat.Sat.Compute()
	return ionoCalc, ndviCalc, radiationCalc, nil
}

func (vcSat *VcSatelite) GetSatelite() *BasicSatelite {
	return vcSat.Sat
}

func (ssSat *SsSatelite) MeasurementTime() time.Duration {
	return ssSat.Sat.MeasurementTime()
}

func (ssSat *SsSatelite) Compute() ([]float64, []float64, []float64, []float64) {
	ionoCalc, ndviCalc, radiationCalc, _ := ssSat.Sat.Compute()
	values := ssSat.SeaSalinities
	min := math.Min(values)
	max := math.Max(values)
	avg := math.Avg(values)
	ssSat.SalinitiesCalc = []float64{min, max, avg}
	return ionoCalc, ndviCalc, radiationCalc, ssSat.SalinitiesCalc[:]
}

func (ssSat *SsSatelite) GetSatelite() *BasicSatelite {
	return ssSat.Sat
}
