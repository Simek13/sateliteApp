package database

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
)

type Satellite struct {
	Id   int    `db:"id" goqu:"skipinsert" goqu:"skipupdate"`
	Name string `db:"name"`
}

const satelliteTable = "satellites"

func (d *MySQLDatabase) AddSatellite(s *Satellite) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}

	return tx.Wrap(func() error {

		_, err := tx.Insert(satelliteTable).
			Prepared(true).
			Rows(s).Executor().
			Exec()
		if err != nil {
			return err
		}
		return nil
	})
}

func (d *MySQLDatabase) GetSatelliteId(name string) (int, error) {
	sql, _, err := d.From(satelliteTable).Select("id").Where(goqu.C("name").Eq(name)).ToSQL()

	if err != nil {
		return -1, errors.Wrap(err, "Error generating sql")
	}
	rows, err := d.Query(sql)
	if err != nil {
		return -1, errors.Wrap(err, "Error executing sql query")
	}
	defer rows.Close()
	var idSat int
	for rows.Next() {
		err := rows.Scan(&idSat)
		if err != nil {
			return -1, errors.Wrap(err, "Error scanning rows")
		}
	}
	err = rows.Err()
	if err != nil {
		return -1, errors.Wrap(err, "Error scanning rows")
	}

	return idSat, nil
}
