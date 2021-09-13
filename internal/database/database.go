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
