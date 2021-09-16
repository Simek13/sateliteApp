package main

import (
	"database/sql"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Simek13/satelliteApp/internal/app"
	"github.com/Simek13/satelliteApp/internal/csv"
	"github.com/Simek13/satelliteApp/internal/database"

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

	dbBaseUrl := fmt.Sprintf("%s:%s@tcp(%s:%s)/", cfg.dbUser, cfg.dbPass, cfg.dbHost, cfg.dbPort)
	// create db
	/* db, err := database.Create(dbBaseUrl, cfg.dbName, cfg.dbType)
	if err != nil {
		fmt.Println("error in db while checking: %w", err)
	} */

	dbUrl := dbBaseUrl + cfg.dbName
	db, err := sql.Open(cfg.dbType, dbUrl)
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error opening database")
	}
	gdb := goqu.New(cfg.dbType, db)
	mysqlDb := &database.MySQLDatabase{gdb}

	fmt.Println("Successfully Connected to MySQL database")
	defer db.Close()

	app.Run(filename, sats, mysqlDb)

}
