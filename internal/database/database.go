package database

import (
	"database/sql"
)

func Create(dbBaseUrl, name, dbType string) (*sql.DB, error) {
	db, err := sql.Open(dbType, dbBaseUrl)
	if err != nil {
		return nil, err
	}

	defer db.Close()

	_, err = db.Exec("CREATE DATABASE " + name)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("USE " + name)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateTables(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE satelites ( id int, name varchar(32))")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE measurements ( filename varchar(32), idSat int, timestamp varchar(32), ionoIndex float, ndviIndex float, radiationIndex float, specificMeasurement varchar(32))")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE TABLE computationResults ( idSat int, duration varchar(32), maxIono float, minIono float, avgIono float, maxNdvi float, minNdvi float, avgNdvi float, maxRad float, minRad float, avgRad float, maxSpec float, minSpec float, avgSpec float)")
	if err != nil {
		return err
	}
	return nil
}
