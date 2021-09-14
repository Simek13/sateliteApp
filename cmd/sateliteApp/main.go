package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Simek13/sateliteApp/internal/csvreader"
	"github.com/Simek13/sateliteApp/internal/satelites"
	"github.com/Simek13/sateliteApp/internal/sort"

	"github.com/doug-martin/goqu/v9"
	"github.com/go-sql-driver/mysql"
	"github.com/namsral/flag"
	"github.com/pkg/errors"

	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
)

// TODO napravi cfg strukturu kao u sunspotu i koristi flagove za neke ulazne parametre kao ime baze i ip adresa, lokacija filea, itd...
var cfg struct {
	inputCsvUrl string
	dateLayout  string

	// DB flags
	dbType string
	dbUser string
	dbPass string
	dbHost string
	dbPort string
	dbName string
}

const DuplicateEntryNum = 1062

func validate() (err error) {
	if cfg.dbType != "mysql" && cfg.dbType != "postgres" && cfg.dbType != "sqlite3" && cfg.dbType != "sqlserver" {
		return errors.New("Invalid db type")
	}
	if len(cfg.dbUser) < 4 || len(cfg.dbUser) > 100 {
		return errors.New("db user is not between 4 and 100 characters")
	}
	if len(cfg.dbPass) < 4 || len(cfg.dbPass) > 100 {
		return errors.New("db password is not between 4 and 100 characters")
	}
	if len(cfg.dbHost) < 4 || len(cfg.dbHost) > 100 {
		return errors.New("db host is not between 4 and 100 characters")
	}
	var port int
	port, err = strconv.Atoi(cfg.dbPort)
	if err != nil {
		port, err = net.LookupPort("tcp", cfg.dbPort)
		if err != nil {
			return errors.New("Invalid db port")
		}
	}
	if port < 1024 || port > 65535 {
		return errors.New("db port is not between 1024 and 65535")
	}
	if len(cfg.dbName) < 4 || len(cfg.dbName) > 100 {
		return errors.New("db name is not between 4 and 100 characters")
	}

	return nil
}

func main() {
	flag.StringVar(&cfg.inputCsvUrl, "url", "https://raw.githubusercontent.com/sea43d/PythonEvaluation/master/satDataCSV2.csv", "url of input csv file")
	flag.StringVar(&cfg.dateLayout, "date_layout", "01-02-2006 15:04", "date layout in csv input file")
	flag.StringVar(&cfg.dbType, "db_type", "mysql", "type of database")
	flag.StringVar(&cfg.dbUser, "db_user", "root", "user name for database")
	flag.StringVar(&cfg.dbPass, "db_pass", "emis", "user password for database")
	flag.StringVar(&cfg.dbHost, "db_host", "127.0.0.1", "host for database")
	flag.StringVar(&cfg.dbPort, "db_port", "3306", "port for database connection")
	flag.StringVar(&cfg.dbName, "db_name", "satelites", "name of database")

	err := validate()
	if err != nil {
		fmt.Println("%w", err)
		os.Exit(1)
	}

	split := strings.Split(cfg.inputCsvUrl, "/")
	filename := split[len(split)-1]
	data, err := csvreader.ReadCsvFromUrl(cfg.inputCsvUrl)
	if err != nil {
		fmt.Println("Error reading csv. %w", err)
		os.Exit(1)
	}

	sats := make(map[string]satelites.Satellite)

	for _, row := range data[1:] {

		satId := row[0]
		if _, ok := sats[satId]; !ok {
			sat := satelites.BasicSatellite{Id: satId,
				Timestamps:       make([]time.Time, 0),
				IonoIndexes:      make([]float64, 0),
				NdviIndexes:      make([]float64, 0),
				RadiationIndexes: make([]float64, 0),
			}
			switch satId {
			case "30J14":
				sat.SatelliteType = satelites.Ea
				sats[satId] = &satelites.EaSatellite{
					BasicSatellite: sat,
					Altitudes:      make([]float64, 0),
				}
			case "13A14", "6N14":
				sat.SatelliteType = satelites.Ss
				sats[satId] = &satelites.SsSatellite{
					BasicSatellite: sat,
					SeaSalinities:  make([]float64, 0),
				}
			case "8J14":
				sat.SatelliteType = satelites.Vc
				sats[satId] = &satelites.VcSatellite{
					BasicSatellite: sat,
					Vegetations:    make([]string, 0),
				}
			}
		}

		tm, err := time.Parse(cfg.dateLayout, row[1])
		if err != nil {
			fmt.Println("Error parsing time. %w", err)
			os.Exit(1)
		}

		satellite := sats[satId]

		satellite.GetSatellite().Timestamps = append(satellite.GetSatellite().Timestamps, tm)

		if val, err := strconv.ParseFloat(row[2], 64); err == nil {
			satellite.GetSatellite().IonoIndexes = append(satellite.GetSatellite().IonoIndexes, val)
		}
		if val, err := strconv.ParseFloat(row[3], 64); err == nil {
			satellite.GetSatellite().NdviIndexes = append(satellite.GetSatellite().NdviIndexes, val)
		}
		if val, err := strconv.ParseFloat(row[4], 64); err == nil {
			satellite.GetSatellite().RadiationIndexes = append(satellite.GetSatellite().RadiationIndexes, val)
		}

		switch satellite.GetSatellite().SatelliteType {
		case satelites.Ea:
			if val, err := strconv.ParseFloat(row[5], 64); err == nil {
				satellite.(*satelites.EaSatellite).Altitudes = append(satellite.(*satelites.EaSatellite).Altitudes, val)
			}
		case satelites.Ss:
			if val, err := strconv.ParseFloat(row[5], 64); err == nil {
				satellite.(*satelites.SsSatellite).SeaSalinities = append(satellite.(*satelites.SsSatellite).SeaSalinities, val)
			}
		case satelites.Vc:
			satellite.(*satelites.VcSatellite).Vegetations = append(satellite.(*satelites.VcSatellite).Vegetations, row[5])
		}
	}

	// Date
	for _, sat := range sats {
		fmt.Println(sat.GetSatellite().Id, "-", sat.MeasurementTime())
	}

	fmt.Println()

	ionoAvg := make(map[string]float64)
	NDVIAvg := make(map[string]float64)
	radAvg := make(map[string]float64)
	altAvg := make(map[string]float64)
	salAvg := make(map[string]float64)

	for id, sat := range sats {

		fmt.Println("Satellite: ", id)
		ionoCalc, ndviCalc, radCalc, specCalc := sat.Compute()
		fmt.Println("Ionosphere index:", ionoCalc[0], "(MIN)", ionoCalc[1], "(MAX)", ionoCalc[2], "(AVG)")
		ionoAvg[id] = ionoCalc[2]
		fmt.Println("NDVI index:", ndviCalc[0], "(MIN)", ndviCalc[1], "(MAX)", ndviCalc[2], "(AVG)")
		NDVIAvg[id] = ndviCalc[2]
		fmt.Println("Radiation index:", radCalc[0], "(MIN)", radCalc[1], "(MAX)", radCalc[2], "(AVG)")
		radAvg[id] = radCalc[2]
		switch sat.GetSatellite().SatelliteType {
		case satelites.Ea:
			fmt.Println("Earth altitude:", specCalc[0], "(MIN)", specCalc[1], "(MAX)", specCalc[2], "(AVG)")
			altAvg[id] = specCalc[2]
		case satelites.Ss:
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

	dbBaseUrl := fmt.Sprintf("%s:%s@tcp(%s:%s)/", cfg.dbUser, cfg.dbPass, cfg.dbHost, cfg.dbPort)
	// create db
	/* db, err := database.Create(dbBaseUrl, cfg.dbName, cfg.dbType)
	if err != nil {
		fmt.Println("error in db while checking: %w", err))
	} */

	// open database
	dbUrl := dbBaseUrl + cfg.dbName
	db, err := sql.Open(cfg.dbType, dbUrl)

	if err != nil {
		fmt.Println("Error opening database. %w", err)
		os.Exit(1)
	}

	defer db.Close()

	fmt.Println("Successfully Connected to MySQL database")

	dialect := goqu.Dialect(cfg.dbType)

	for name := range sats {
		sql, _, err := dialect.Insert("satelites").Cols("name").Vals(goqu.Vals{name}).ToSQL()

		if err != nil {
			fmt.Println("Error generating sql. %w", err)
			os.Exit(1)
		}

		_, err = db.Exec(sql)

		if err != nil {
			if err.(*mysql.MySQLError).Number != DuplicateEntryNum {
				fmt.Println("Error executing sql query. %w", err)
				os.Exit(1)
			}
		}
	}

	for _, row := range data[1:] {

		sql, _, err := dialect.From("satelites").Select("id").Where(goqu.C("name").Eq(row[0])).ToSQL()

		if err != nil {
			fmt.Println("Error generating sql. %w", err)
			os.Exit(1)
		}
		rows, err := db.Query(sql)
		if err != nil {
			fmt.Println("Error executing sql query. %w", err)
			os.Exit(1)
		}
		defer rows.Close()
		var idSat int
		for rows.Next() {
			err := rows.Scan(&idSat)
			if err != nil {
				fmt.Println("Error scanning rows. %w", err)
				os.Exit(1)
			}
		}
		err = rows.Err()
		if err != nil {
			fmt.Println("Error scanning rows. %w", err)
			os.Exit(1)
		}

		ionoIndex, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			fmt.Println("Cannot parse given value to float. %w", err)
			os.Exit(1)
		}
		ndviIndex, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			fmt.Println("Cannot parse given value to float. %w", err)
			os.Exit(1)
		}
		radiationIndex, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			fmt.Println("Cannot parse given value to float. %w", err)
			os.Exit(1)
		}

		sql, _, err = dialect.Insert("measurements").
			Cols("filename", "idSat", "timestamp", "ionoIndex", "ndviIndex", "radiationIndex", "specificMeasurement").
			Vals(goqu.Vals{filename, idSat, row[1], ionoIndex, ndviIndex, radiationIndex, row[5]}).ToSQL()

		if err != nil {
			fmt.Println("Error generating sql. %w", err)
			os.Exit(1)
		}
		_, err = db.Exec(sql)

		if err != nil {
			if err.(*mysql.MySQLError).Number != DuplicateEntryNum {
				fmt.Println("Error executing sql query. %w", err)
				os.Exit(1)
			}
		}
	}

	for name, sat := range sats {
		sql, _, err := dialect.From("satelites").Select("id").Where(goqu.C("name").Eq(name)).ToSQL()
		if err != nil {
			fmt.Println("Error generating sql. %w", err)
			os.Exit(1)
		}
		rows, err := db.Query(sql)
		if err != nil {
			fmt.Println("Error executing sql query. %w", err)
			os.Exit(1)
		}
		defer rows.Close()
		var idSat int
		for rows.Next() {
			err := rows.Scan(&idSat)
			if err != nil {
				fmt.Println("Error scanning rows. %w", err)
				os.Exit(1)
			}
		}
		err = rows.Err()
		if err != nil {
			fmt.Println("Error scanning rows. %w", err)
			os.Exit(1)
		}
		bSat := sat.GetSatellite()
		var ds *goqu.InsertDataset
		switch sat.GetSatellite().SatelliteType {
		case satelites.Ea:
			ds = dialect.Insert("computationResults").
				Cols("idSat", "duration", "maxIono", "minIono", "avgIono", "maxNdvi", "minNdvi", "avgNdvi", "maxRad", "minRad", "avgRad", "maxSpec", "minSpec", "avgSpec").
				Vals(goqu.Vals{idSat, fmt.Sprint(bSat.Duration), bSat.IonoCalc[0], bSat.IonoCalc[1], bSat.IonoCalc[2], bSat.NdviCalc[0], bSat.NdviCalc[1], bSat.NdviCalc[2], bSat.RadiationCalc[0], bSat.RadiationCalc[1], bSat.RadiationCalc[2], sat.(*satelites.EaSatellite).AltitudesCalc[0], sat.(*satelites.EaSatellite).AltitudesCalc[1], sat.(*satelites.EaSatellite).AltitudesCalc[2]})
		case satelites.Ss:
			ds = dialect.Insert("computationResults").
				Cols("idSat", "duration", "maxIono", "minIono", "avgIono", "maxNdvi", "minNdvi", "avgNdvi", "maxRad", "minRad", "avgRad", "maxSpec", "minSpec", "avgSpec").
				Vals(goqu.Vals{idSat, fmt.Sprint(bSat.Duration), bSat.IonoCalc[0], bSat.IonoCalc[1], bSat.IonoCalc[2], bSat.NdviCalc[0], bSat.NdviCalc[1], bSat.NdviCalc[2], bSat.RadiationCalc[0], bSat.RadiationCalc[1], bSat.RadiationCalc[2], sat.(*satelites.SsSatellite).SalinitiesCalc[0], sat.(*satelites.SsSatellite).SalinitiesCalc[1], sat.(*satelites.SsSatellite).SalinitiesCalc[2]})
		case satelites.Vc:
			ds = dialect.Insert("computationResults").
				Cols("idSat", "duration", "maxIono", "minIono", "avgIono", "maxNdvi", "minNdvi", "avgNdvi", "maxRad", "minRad", "avgRad", "maxSpec", "minSpec", "avgSpec").
				Vals(goqu.Vals{idSat, fmt.Sprint(bSat.Duration), bSat.IonoCalc[0], bSat.IonoCalc[1], bSat.IonoCalc[2], bSat.NdviCalc[0], bSat.NdviCalc[1], bSat.NdviCalc[2], bSat.RadiationCalc[0], bSat.RadiationCalc[1], bSat.RadiationCalc[2], 0, 0, 0})
		}

		sql, _, err = ds.ToSQL()

		if err != nil {
			fmt.Println("Error generating sql. %w", err)
			os.Exit(1)
		}
		_, err = db.Exec(sql)

		if err != nil {
			if err.(*mysql.MySQLError).Number != DuplicateEntryNum {
				fmt.Println("Error executing sql query. %w", err)
				os.Exit(1)
			}
		}
	}

}
