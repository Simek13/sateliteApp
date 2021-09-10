package main

import (
	"database/sql"
	"fmt"
	"satelites/csvreader"
	"satelites/math"
	"satelites/sort"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"

	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/go-sql-driver/mysql"
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
	duration         time.Duration
	ionoCalc         []float64
	ndviCalc         []float64
	radiationCalc    []float64
}

type EaSatelite struct {
	sat           *BasicSatelite
	altitudes     []float64
	altitudesCalc []float64
}

type VcSatelite struct {
	sat         *BasicSatelite
	vegetations []string
}

type SsSatelite struct {
	sat            *BasicSatelite
	seaSalinities  []float64
	salinitiesCalc []float64
}

func (sat *BasicSatelite) measurementTime() time.Duration {
	minD := math.MinDate(sat.timestamps)
	maxD := math.MaxDate(sat.timestamps)
	timeDiff := maxD.Sub(minD)
	sat.duration = timeDiff
	return timeDiff
}

func (sat *BasicSatelite) compute() ([]float64, []float64, []float64, []float64) {
	values := sat.ionoIndexes
	min := math.Min(values)
	max := math.Max(values)
	avg := math.Avg(values)
	sat.ionoCalc = []float64{min, max, avg}
	values = sat.ndviIndexes
	min = math.Min(values)
	max = math.Max(values)
	avg = math.Avg(values)
	sat.ndviCalc = []float64{min, max, avg}
	values = sat.radiationIndexes
	min = math.Min(values)
	max = math.Max(values)
	avg = math.Avg(values)
	sat.radiationCalc = []float64{min, max, avg}

	return sat.ionoCalc[:], sat.ndviCalc[:], sat.radiationCalc[:], nil

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
	eaSat.altitudesCalc = []float64{min, max, avg}
	return ionoCalc, ndviCalc, radiationCalc, eaSat.altitudesCalc[:]
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
	ssSat.salinitiesCalc = []float64{min, max, avg}
	return ionoCalc, ndviCalc, radiationCalc, ssSat.salinitiesCalc[:]
}

func (ssSat *SsSatelite) getSatelite() *BasicSatelite {
	return ssSat.sat
}

func main() {
	url := "https://raw.githubusercontent.com/sea43d/PythonEvaluation/master/satDataCSV2.csv"
	split := strings.Split(url, "/")
	filename := split[len(split)-1]
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
			panic(err)
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

	// store data to db

	// create db
	/* db, err := database.Create("satelites")
	if err != nil {
		panic(err)
	} */

	// open database
	db, err := sql.Open("mysql", "root:emis@tcp(127.0.0.1:3306)/satelites")

	if err != nil {
		panic(err)
	}

	// defer db.Close()

	fmt.Println("Successfully Connected to MySQL database")

	_, err = db.Exec("CREATE TABLE satelites ( id int, name varchar(32))")
	if err != nil {
		//panic(err)
	}

	_, err = db.Exec("CREATE TABLE measurements ( filename varchar(32), idSat int, timestamp varchar(32), ionoIndex float, ndviIndex float, radiationIndex float, specificMeasurement varchar(32))")
	if err != nil {
		//panic(err)
	}
	_, err = db.Exec("CREATE TABLE computationResults ( idSat int, duration varchar(32), maxIono float, minIono float, avgIono float, maxNdvi float, minNdvi float, avgNdvi float, maxRad float, minRad float, avgRad float, maxSpec float, minSpec float, avgSpec float)")
	if err != nil {
		//panic(err)
	}

	dialect := goqu.Dialect("mysql")

	db_satelites := make(map[string]int)
	i := 1
	for id := range satelites {
		db_satelites[id] = i
		ds := dialect.Insert("satelites").Cols("id", "name").Vals(goqu.Vals{i, id})
		sql, _, _ := ds.ToSQL()

		_, err = db.Exec(sql)

		if err != nil {
			panic(err)
		}
		i++
	}

	for _, row := range data[1:] {
		ionoIndex, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			panic(err)
		}
		ndviIndex, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			panic(err)
		}
		radiationIndex, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			panic(err)
		}

		ds := dialect.Insert("measurements").
			Cols("filename", "idSat", "timestamp", "ionoIndex", "ndviIndex", "radiationIndex", "specificMeasurement").
			Vals(goqu.Vals{filename, db_satelites[row[0]], row[1], ionoIndex, ndviIndex, radiationIndex, row[5]})

		sql, _, _ := ds.ToSQL()

		_, err = db.Exec(sql)

		if err != nil {
			panic(err)
		}
	}

	for id, sat := range satelites {
		bSat := sat.getSatelite()
		var ds *goqu.InsertDataset
		switch sat.(type) {
		case *EaSatelite:
			ds = dialect.Insert("computationResults").
				Cols("idSat", "duration", "maxIono", "minIono", "avgIono", "maxNdvi", "minNdvi", "avgNdvi", "maxRad", "minRad", "avgRad", "maxSpec", "minSpec", "avgSpec").
				Vals(goqu.Vals{db_satelites[id], fmt.Sprint(bSat.duration), bSat.ionoCalc[0], bSat.ionoCalc[1], bSat.ionoCalc[2], bSat.ndviCalc[0], bSat.ndviCalc[1], bSat.ndviCalc[2], bSat.radiationCalc[0], bSat.radiationCalc[1], bSat.radiationCalc[2], sat.(*EaSatelite).altitudesCalc[0], sat.(*EaSatelite).altitudesCalc[1], sat.(*EaSatelite).altitudesCalc[2]})
		case *SsSatelite:
			ds = dialect.Insert("computationResults").
				Cols("idSat", "duration", "maxIono", "minIono", "avgIono", "maxNdvi", "minNdvi", "avgNdvi", "maxRad", "minRad", "avgRad", "maxSpec", "minSpec", "avgSpec").
				Vals(goqu.Vals{db_satelites[id], fmt.Sprint(bSat.duration), bSat.ionoCalc[0], bSat.ionoCalc[1], bSat.ionoCalc[2], bSat.ndviCalc[0], bSat.ndviCalc[1], bSat.ndviCalc[2], bSat.radiationCalc[0], bSat.radiationCalc[1], bSat.radiationCalc[2], sat.(*SsSatelite).salinitiesCalc[0], sat.(*SsSatelite).salinitiesCalc[1], sat.(*SsSatelite).salinitiesCalc[2]})
		case *VcSatelite:
			ds = dialect.Insert("computationResults").
				Cols("idSat", "duration", "maxIono", "minIono", "avgIono", "maxNdvi", "minNdvi", "avgNdvi", "maxRad", "minRad", "avgRad", "maxSpec", "minSpec", "avgSpec").
				Vals(goqu.Vals{db_satelites[id], fmt.Sprint(bSat.duration), bSat.ionoCalc[0], bSat.ionoCalc[1], bSat.ionoCalc[2], bSat.ndviCalc[0], bSat.ndviCalc[1], bSat.ndviCalc[2], bSat.radiationCalc[0], bSat.radiationCalc[1], bSat.radiationCalc[2], 0, 0, 0})
		}

		sql, _, _ := ds.ToSQL()

		_, err = db.Exec(sql)

		if err != nil {
			panic(err)
		}
	}

	db.Close()

}
