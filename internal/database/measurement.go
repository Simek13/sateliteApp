package database

type Measurement struct {
	Id                  int     `db:"id" goqu:"skipinsert, skipupdate"`
	FileName            string  `db:"filename"`
	IdSat               int     `db:"idSat"`
	Timestamp           string  `db:"timestamp"`
	IonoIndex           float64 `db:"ionoIndex"`
	NdviIndex           float64 `db:"ndviIndex"`
	RadiationIndex      float64 `db:"radiationIndex"`
	SpecificMeasurement string  `db:"specificMeasurement"`
}

const measurementTable = "measurements"

func (d *MySQLDatabase) AddMeasurement(m *Measurement) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}

	return tx.Wrap(func() error {

		_, err := tx.Insert(measurementTable).
			Prepared(true).
			Rows(m).Executor().
			Exec()
		if err != nil {
			return err
		}
		return nil
	})
}
