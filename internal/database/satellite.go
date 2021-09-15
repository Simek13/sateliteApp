package database

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
