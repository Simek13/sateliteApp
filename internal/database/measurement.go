package database

import (
	"fmt"
	"time"

	"github.com/Simek13/satelliteApp/internal/satellites"
	"github.com/pkg/errors"
)

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

func (d *MySQLDatabase) AddMeasurements(filename string, sats map[string]satellites.Satellite) error {
	for name, sat := range sats {

		bSat := sat.GetSatellite()
		idSat, err := d.GetSatelliteId(name)
		if err != nil {
			return err
		}
		var timestamp time.Time
		var ionoIndex, ndviIndex, radiationIndex float64
		for i := range sat.GetSatellite().Timestamps {
			timestamp = bSat.Timestamps[i]
			ionoIndex = bSat.IonoIndexes[i]
			ndviIndex = bSat.NdviIndexes[i]
			radiationIndex = bSat.RadiationIndexes[i]

			m := &Measurement{
				FileName:       filename,
				IdSat:          idSat,
				Timestamp:      timestamp.String(),
				IonoIndex:      ionoIndex,
				NdviIndex:      ndviIndex,
				RadiationIndex: radiationIndex,
			}
			switch s := sat.(type) {
			case *satellites.EaSatellite:
				m.SpecificMeasurement = fmt.Sprintf("%f", s.Altitudes[i])
			case *satellites.SsSatellite:
				m.SpecificMeasurement = fmt.Sprintf("%f", s.SeaSalinities[i])
			case *satellites.VcSatellite:
				m.SpecificMeasurement = s.Vegetations[i]
			}

			err = d.AddMeasurement(m)

			err = HandleSqlError(err)
			if err != nil {
				return errors.Wrap(err, "Unable to insert measurement into database")
			}
		}

	}
	return nil
}
