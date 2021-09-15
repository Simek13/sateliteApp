package main

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Simek13/satelliteApp/internal/csv"
	"github.com/Simek13/satelliteApp/internal/database"
	"github.com/Simek13/satelliteApp/internal/satellites"
	"github.com/Simek13/satelliteApp/internal/sort"

	"github.com/doug-martin/goqu/v9"
	"github.com/namsral/flag"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
)

// TODO napravi cfg strukturu kao u sunspotu i koristi flagove za neke ulazne parametre kao ime baze i ip adresa, lokacija filea, itd...
var cfg struct {
	inputCsvUrl string

	// DB flags
	dbType string
	dbUser string
	dbPass string
	dbHost string
	dbPort string
	dbName string
}

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
	ctxlog := log.WithFields(log.Fields{"event": "main"})
	flag.StringVar(&cfg.inputCsvUrl, "url", "https://raw.githubusercontent.com/sea43d/PythonEvaluation/master/satDataCSV2.csv", "url of input csv file")
	flag.StringVar(&cfg.dbType, "db_type", "mysql", "type of database")
	flag.StringVar(&cfg.dbUser, "db_user", "root", "user name for database")
	flag.StringVar(&cfg.dbPass, "db_pass", "emis", "user password for database")
	flag.StringVar(&cfg.dbHost, "db_host", "127.0.0.1", "host for database")
	flag.StringVar(&cfg.dbPort, "db_port", "3306", "port for database connection")
	flag.StringVar(&cfg.dbName, "db_name", "satellites", "name of database")

	err := validate()
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal(err)
	}

	split := strings.Split(cfg.inputCsvUrl, "/")
	filename := split[len(split)-1]
	data, err := csv.ReadCsvFromUrl(cfg.inputCsvUrl)
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error reading csv")
	}

	sats, err := csv.ParseCsvData(data)
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error parsing csv data")
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

	// kako ovo raščlaniti na funkcije koje rade samo jednu stvar
	for id, sat := range sats {

		fmt.Println("Satellite: ", id)
		ionoCalc, ndviCalc, radCalc, specCalc := sat.Compute()
		fmt.Println("Ionosphere index:", ionoCalc[0], "(MIN)", ionoCalc[1], "(MAX)", ionoCalc[2], "(AVG)")
		ionoAvg[id] = ionoCalc[2]
		fmt.Println("NDVI index:", ndviCalc[0], "(MIN)", ndviCalc[1], "(MAX)", ndviCalc[2], "(AVG)")
		NDVIAvg[id] = ndviCalc[2]
		fmt.Println("Radiation index:", radCalc[0], "(MIN)", radCalc[1], "(MAX)", radCalc[2], "(AVG)")
		radAvg[id] = radCalc[2]
		switch sat.(type) {
		case *satellites.EaSatellite:
			fmt.Println("Earth altitude:", specCalc[0], "(MIN)", specCalc[1], "(MAX)", specCalc[2], "(AVG)")
			altAvg[id] = specCalc[2]
		case *satellites.SsSatellite:
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
		fmt.Println("error in db while checking: %w", err)
	} */

	// open database
	dbUrl := dbBaseUrl + cfg.dbName
	db, err := sql.Open(cfg.dbType, dbUrl)
	gdb := goqu.New(cfg.dbType, db)
	mysqlDb := &database.MySQLDatabase{gdb}

	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error opening database")
	}

	defer db.Close()

	fmt.Println("Successfully Connected to MySQL database")

	for name := range sats {
		s := &database.Satellite{Name: name}
		err := mysqlDb.AddSatellite(s)

		err = database.HandleSqlError(err)
		if err != nil {
			ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Unable to insert satellite into database")
		}
	}

	for _, row := range data[1:] {

		idSat, err := mysqlDb.GetSatelliteId(row[0])
		if err != nil {
			ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal(errors.Unwrap(err))
		}

		ionoIndex, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Cannot parse given value to float")
		}
		ndviIndex, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Cannot parse given value to float")
		}
		radiationIndex, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Cannot parse given value to float")
		}

		m := &database.Measurement{
			FileName:            filename,
			IdSat:               idSat,
			Timestamp:           row[1],
			IonoIndex:           ionoIndex,
			NdviIndex:           ndviIndex,
			RadiationIndex:      radiationIndex,
			SpecificMeasurement: row[5],
		}
		err = mysqlDb.AddMeasurement(m)

		err = database.HandleSqlError(err)
		if err != nil {
			ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Unable to insert measurement into database")
		}
	}

	for name, sat := range sats {
		idSat, err := mysqlDb.GetSatelliteId(name)
		if err != nil {
			ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal(errors.Unwrap(err))
		}
		bSat := sat.GetSatellite()
		c := &database.Computation{
			IdSat:    idSat,
			Duration: fmt.Sprint(bSat.Duration),
			MaxIono:  bSat.IonoCalc[0],
			MinIono:  bSat.IonoCalc[1],
			AvgIono:  bSat.IonoCalc[2],
			MaxNdvi:  bSat.NdviCalc[0],
			MinNdvi:  bSat.NdviCalc[1],
			AvgNdvi:  bSat.NdviCalc[2],
			MaxRad:   bSat.RadiationCalc[0],
			MinRad:   bSat.RadiationCalc[1],
			AvgRad:   bSat.RadiationCalc[2],
		}
		switch s := sat.(type) {
		case *satellites.EaSatellite:
			c.MaxSpec = s.AltitudesCalc[0]
			c.MinSpec = s.AltitudesCalc[1]
			c.AvgSpec = s.AltitudesCalc[2]
		case *satellites.SsSatellite:
			c.MaxSpec = s.SalinitiesCalc[0]
			c.MinSpec = s.SalinitiesCalc[1]
			c.AvgSpec = s.SalinitiesCalc[2]
		case *satellites.VcSatellite:
		}
		err = mysqlDb.AddComputations(c)

		err = database.HandleSqlError(err)
		if err != nil {
			ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Unable to insert computation into database")
		}
	}

}
