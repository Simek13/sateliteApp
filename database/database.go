package database

import (
	"database/sql"
)

func Create(name string) (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:emis@tcp(127.0.0.1:3306)/")
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
