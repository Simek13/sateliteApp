package database

import (
	"fmt"
	"time"

	"github.com/Simek13/satelliteApp/internal/satellites"
	pb "github.com/Simek13/satelliteApp/pkg"
	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
)

const measurementTable = "measurements"

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

func (m *Measurement) Protobuf() *pb.Measurement {
	if m == nil {
		return nil
	}

	measurement := &pb.Measurement{
		Id:                  int32(m.Id),
		FileName:            m.FileName,
		IdSat:               int32(m.IdSat),
		Timestamp:           m.Timestamp,
		IonoIndex:           float32(m.IonoIndex),
		NdviIndex:           float32(m.NdviIndex),
		RadiationIndex:      float32(m.RadiationIndex),
		SpecificMeasurement: m.SpecificMeasurement,
	}
	return measurement
}

func NewMeasurement(m *pb.Measurement) *Measurement {
	if m == nil {
		return nil
	}
	measurement := &Measurement{
		Id:                  int(m.Id),
		FileName:            m.FileName,
		IdSat:               int(m.IdSat),
		Timestamp:           m.Timestamp,
		IonoIndex:           float64(m.IonoIndex),
		NdviIndex:           float64(m.NdviIndex),
		RadiationIndex:      float64(m.RadiationIndex),
		SpecificMeasurement: m.SpecificMeasurement,
	}
	return measurement
}

func (m Measurement) String() string {
	return fmt.Sprintf("Id: %v, Filename: %s, IdSat: %v, Timestamp: %s, IonoIndex: %v, NdviIndex: %v, RadiationIndex: %v, SpecificMeasurement: %s",
		m.Id, m.FileName, m.IdSat, m.Timestamp, m.IonoIndex, m.NdviIndex, m.RadiationIndex, m.SpecificMeasurement)
}

func (d *MySQLDatabase) AddMeasurement(m *Measurement) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}
	fmt.Println(m)

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

func (d *MySQLDatabase) GetMeasurements(satId int) ([]Measurement, error) {
	var sql string
	var err error
	if satId != 0 {
		sql, _, err = d.From(measurementTable).Where(goqu.C("idSat").Eq(satId)).ToSQL()
	} else {
		sql, _, err = d.From(measurementTable).ToSQL()
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error generating sql")
	}
	rows, err := d.Query(sql)
	if err != nil {
		return nil, errors.Wrap(err, "Error executing sql query")
	}
	defer rows.Close()
	measurements := make([]Measurement, 0)
	for rows.Next() {
		var m Measurement
		err := rows.Scan(&m.Id, &m.FileName, &m.IdSat, &m.Timestamp, &m.IonoIndex, &m.NdviIndex, &m.RadiationIndex, &m.SpecificMeasurement)
		if err != nil {
			return nil, errors.Wrap(err, "Error scanning rows")
		}
		measurements = append(measurements, m)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "Error scanning rows")
	}

	return measurements, nil
}
