package main

import (
	"database/sql"
	"fmt"
	"sateliteApp/internal/csvreader"
	"sateliteApp/internal/satelites"
	"sateliteApp/internal/sort"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"

	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	url := "https://raw.githubusercontent.com/sea43d/PythonEvaluation/master/satDataCSV2.csv"
	split := strings.Split(url, "/")
	filename := split[len(split)-1]
	data, err := csvreader.ReadCsvFromUrl(url)
	if err != nil {
		panic(err)
	}

	sats := make(map[string]satelites.Satelite)

	for _, row := range data[1:] {

		satId := row[0]
		if _, ok := sats[satId]; !ok {
			sat := &satelites.BasicSatelite{Id: satId,
				Timestamps:       make([]time.Time, 0),
				IonoIndexes:      make([]float64, 0),
				NdviIndexes:      make([]float64, 0),
				RadiationIndexes: make([]float64, 0),
			}
			switch satId {
			case "30J14":
				sats[satId] = &satelites.EaSatelite{
					Sat:       sat,
					Altitudes: make([]float64, 0),
				}
			case "13A14", "6N14":
				sats[satId] = &satelites.SsSatelite{
					Sat:           sat,
					SeaSalinities: make([]float64, 0),
				}
			default:
				sats[satId] = &satelites.VcSatelite{
					Sat:         sat,
					Vegetations: make([]string, 0),
				}
			}
		}

		layout := "01-02-2006 15:04"
		tm, err := time.Parse(layout, row[1])
		if err != nil {
			panic(err)
		}

		satelite := sats[satId]

		satelite.GetSatelite().Timestamps = append(satelite.GetSatelite().Timestamps, tm)

		if val, err := strconv.ParseFloat(row[2], 64); err == nil {
			satelite.GetSatelite().IonoIndexes = append(satelite.GetSatelite().IonoIndexes, val)
		}
		if val, err := strconv.ParseFloat(row[3], 64); err == nil {
			satelite.GetSatelite().NdviIndexes = append(satelite.GetSatelite().NdviIndexes, val)
		}
		if val, err := strconv.ParseFloat(row[4], 64); err == nil {
			satelite.GetSatelite().RadiationIndexes = append(satelite.GetSatelite().RadiationIndexes, val)
		}

		switch satelite.(type) {
		case *satelites.EaSatelite:
			if val, err := strconv.ParseFloat(row[5], 64); err == nil {
				satelite.(*satelites.EaSatelite).Altitudes = append(satelite.(*satelites.EaSatelite).Altitudes, val)
			}
		case *satelites.SsSatelite:
			if val, err := strconv.ParseFloat(row[5], 64); err == nil {
				satelite.(*satelites.SsSatelite).SeaSalinities = append(satelite.(*satelites.SsSatelite).SeaSalinities, val)
			}
		case *satelites.VcSatelite:
			satelite.(*satelites.VcSatelite).Vegetations = append(satelite.(*satelites.VcSatelite).Vegetations, row[5])
		}
	}

	// Date
	for _, sat := range sats {
		fmt.Println(sat.GetSatelite().Id, "-", sat.MeasurementTime())
	}

	fmt.Println()

	ionoAvg := make(map[string]float64)
	NDVIAvg := make(map[string]float64)
	radAvg := make(map[string]float64)
	altAvg := make(map[string]float64)
	salAvg := make(map[string]float64)

	for id, sat := range sats {

		fmt.Println("Satelite: ", id)
		ionoCalc, ndviCalc, radCalc, specCalc := sat.Compute()
		fmt.Println("Ionosphere index:", ionoCalc[0], "(MIN)", ionoCalc[1], "(MAX)", ionoCalc[2], "(AVG)")
		ionoAvg[id] = ionoCalc[2]
		fmt.Println("NDVI index:", ndviCalc[0], "(MIN)", ndviCalc[1], "(MAX)", ndviCalc[2], "(AVG)")
		NDVIAvg[id] = ndviCalc[2]
		fmt.Println("Radiation index:", radCalc[0], "(MIN)", radCalc[1], "(MAX)", radCalc[2], "(AVG)")
		radAvg[id] = radCalc[2]
		switch sat.(type) {
		case *satelites.EaSatelite:
			fmt.Println("Earth altitude:", specCalc[0], "(MIN)", specCalc[1], "(MAX)", specCalc[2], "(AVG)")
			altAvg[id] = specCalc[2]
		case *satelites.SsSatelite:
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

	fmt.Println("Successfully Connected to MySQL database")

	//create tables
	/* err = database.CreateTables(db)
	if err != nil {
		panic(err)
	} */

	dialect := goqu.Dialect("mysql")

	db_satelites := make(map[string]int)
	i := 1
	for id := range sats {
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

	for id, sat := range sats {
		bSat := sat.GetSatelite()
		var ds *goqu.InsertDataset
		switch sat.(type) {
		case *satelites.EaSatelite:
			ds = dialect.Insert("computationResults").
				Cols("idSat", "duration", "maxIono", "minIono", "avgIono", "maxNdvi", "minNdvi", "avgNdvi", "maxRad", "minRad", "avgRad", "maxSpec", "minSpec", "avgSpec").
				Vals(goqu.Vals{db_satelites[id], fmt.Sprint(bSat.Duration), bSat.IonoCalc[0], bSat.IonoCalc[1], bSat.IonoCalc[2], bSat.NdviCalc[0], bSat.NdviCalc[1], bSat.NdviCalc[2], bSat.RadiationCalc[0], bSat.RadiationCalc[1], bSat.RadiationCalc[2], sat.(*satelites.EaSatelite).AltitudesCalc[0], sat.(*satelites.EaSatelite).AltitudesCalc[1], sat.(*satelites.EaSatelite).AltitudesCalc[2]})
		case *satelites.SsSatelite:
			ds = dialect.Insert("computationResults").
				Cols("idSat", "duration", "maxIono", "minIono", "avgIono", "maxNdvi", "minNdvi", "avgNdvi", "maxRad", "minRad", "avgRad", "maxSpec", "minSpec", "avgSpec").
				Vals(goqu.Vals{db_satelites[id], fmt.Sprint(bSat.Duration), bSat.IonoCalc[0], bSat.IonoCalc[1], bSat.IonoCalc[2], bSat.NdviCalc[0], bSat.NdviCalc[1], bSat.NdviCalc[2], bSat.RadiationCalc[0], bSat.RadiationCalc[1], bSat.RadiationCalc[2], sat.(*satelites.SsSatelite).SalinitiesCalc[0], sat.(*satelites.SsSatelite).SalinitiesCalc[1], sat.(*satelites.SsSatelite).SalinitiesCalc[2]})
		case *satelites.VcSatelite:
			ds = dialect.Insert("computationResults").
				Cols("idSat", "duration", "maxIono", "minIono", "avgIono", "maxNdvi", "minNdvi", "avgNdvi", "maxRad", "minRad", "avgRad", "maxSpec", "minSpec", "avgSpec").
				Vals(goqu.Vals{db_satelites[id], fmt.Sprint(bSat.Duration), bSat.IonoCalc[0], bSat.IonoCalc[1], bSat.IonoCalc[2], bSat.NdviCalc[0], bSat.NdviCalc[1], bSat.NdviCalc[2], bSat.RadiationCalc[0], bSat.RadiationCalc[1], bSat.RadiationCalc[2], 0, 0, 0})
		}

		sql, _, _ := ds.ToSQL()

		_, err = db.Exec(sql)

		if err != nil {
			panic(err)
		}
	}

	db.Close()

}
