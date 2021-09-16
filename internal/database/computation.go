package database

import (
	"fmt"

	"github.com/Simek13/satelliteApp/internal/satellites"
	"github.com/pkg/errors"
)

type Computation struct {
	Id       int     `db:"id" goqu:"skipinsert, skipupdate"`
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

func (d *MySQLDatabase) AddComputation(c *Computation) error {
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

func (d *MySQLDatabase) AddComputations(sats map[string]satellites.Satellite) error {
	for name, sat := range sats {
		idSat, err := d.GetSatelliteId(name)
		if err != nil {
			return err
		}
		bSat := sat.GetSatellite()
		c := &Computation{
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
		err = d.AddComputation(c)

		err = HandleSqlError(err)
		if err != nil {
			return errors.Wrap(err, "Unable to insert computation into database")
		}
	}
	return nil
}
