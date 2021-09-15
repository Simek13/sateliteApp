package database

import (
	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

const DuplicateEntryNum = 1062

type MySQLDatabase struct {
	*goqu.Database
}

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

func HandleSqlError(err error) error {
	if err != nil {
		if e, ok := err.(*mysql.MySQLError); ok {
			if e.Number != DuplicateEntryNum {
				return errors.Wrap(err, "Error inserting into database")
			}
		} else {
			return errors.Wrap(err, "Error inserting into database")
		}
	}
	return nil
}
