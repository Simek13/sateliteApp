package database

type Computation struct {
	Id       int     `db:"id" goqu:"skipinsert" goqu:"skipupdate"`
	IdSat    int     `db:"idSat"`
	Duration string  `db:"duration"`
	MaxIono  float64 `db:"maxIono"`
	MinIono  float64 `db:"minIono"`
	AvgIono  float64 `db:"avgIono"`
	MaxNdvi  float64 `db:"maxNdvi"`
	MinNdvi  float64 `db:"minNdvi"`
	AvgNdvi  float64 `db:"avgNdvi"`
	MaxRad   float64 `db:"maxRad"`
	MinRad   float64 `db:"minRad"`
	AvgRad   float64 `db:"avgRad"`
	MaxSpec  float64 `db:"maxSpec"`
	MinSpec  float64 `db:"minSpec"`
	AvgSpec  float64 `db:"avgSpec"`
}

const computationTable = "computations"

func (d *MySQLDatabase) AddComputations(c *Computation) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}

	return tx.Wrap(func() error {

		_, err := tx.Insert(computationTable).
			Prepared(true).
			Rows(c).Executor().
			Exec()
		if err != nil {
			return err
		}
		return nil
	})
}
