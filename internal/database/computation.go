package database

import (
	"fmt"

	pb "github.com/Simek13/satelliteApp/internal/satellite_communication"
	"github.com/Simek13/satelliteApp/internal/satellites"
	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"
)

const computationTable = "computations"

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

func (c Computation) String() string {
	return fmt.Sprintf("Id: %v, IdSat: %v, Duration: %s, MaxIono: %v, MinIono: %v, AvgIono: %v, MaxNdvi: %v, MinNdvi: %v, AvgNdvi: %v, MaxRad: %v, MinRad: %v, AvgRad: %v, MaxSpec: %v, MinSpec: %v, AvgSpec: %v",
		c.Id, c.IdSat, c.Duration, c.MaxIono, c.MinIono, c.AvgIono, c.MaxNdvi, c.MinNdvi, c.AvgNdvi, c.MaxRad, c.MinRad, c.AvgRad, c.MaxSpec, c.MinSpec, c.AvgSpec)
}

func (c *Computation) Protobuf() *pb.Computation {
	if c == nil {
		return nil
	}

	computation := &pb.Computation{
		Id:       int32(c.Id),
		IdSat:    int32(c.IdSat),
		Duration: c.Duration,
		MaxIono:  float32(c.MaxIono),
		MinIono:  float32(c.MinIono),
		AvgIono:  float32(c.AvgIono),
		MaxNdvi:  float32(c.MaxNdvi),
		MinNdvi:  float32(c.MinNdvi),
		AvgNdvi:  float32(c.AvgNdvi),
		MaxRad:   float32(c.MaxRad),
		MinRad:   float32(c.MinRad),
		AvgRad:   float32(c.AvgRad),
		MaxSpec:  float32(c.MaxSpec),
		MinSpec:  float32(c.MinSpec),
		AvgSpec:  float32(c.AvgSpec),
	}
	return computation
}

func NewComputation(c *pb.Computation) *Computation {
	if c == nil {
		return nil
	}
	computation := &Computation{
		Id:       int(c.Id),
		IdSat:    int(c.IdSat),
		Duration: c.Duration,
		MaxIono:  float64(c.MaxIono),
		MinIono:  float64(c.MinIono),
		AvgIono:  float64(c.AvgIono),
		MaxNdvi:  float64(c.MaxNdvi),
		MinNdvi:  float64(c.MinNdvi),
		AvgNdvi:  float64(c.AvgNdvi),
		MaxRad:   float64(c.MaxRad),
		MinRad:   float64(c.MinRad),
		AvgRad:   float64(c.AvgRad),
		MaxSpec:  float64(c.MaxSpec),
		MinSpec:  float64(c.MinSpec),
		AvgSpec:  float64(c.AvgSpec),
	}
	return computation
}

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

func (d *MySQLDatabase) GetComputations(satName string) ([]Computation, error) {
	var satId int
	var sql string
	var err error
	if satName != "" {
		satId, err = d.GetSatelliteId(satName)
		if err != nil {
			return nil, errors.Wrap(err, "Invalid satellite name")
		}
		sql, _, err = d.From(computationTable).Where(goqu.C("idSat").Eq(satId)).ToSQL()
	} else {
		sql, _, err = d.From(computationTable).ToSQL()
	}

	if err != nil {
		return nil, errors.Wrap(err, "Error generating sql")
	}
	rows, err := d.Query(sql)
	if err != nil {
		return nil, errors.Wrap(err, "Error executing sql query")
	}
	defer rows.Close()
	computations := make([]Computation, 0)
	for rows.Next() {
		var c Computation
		err := rows.Scan(&c.Id, &c.IdSat, &c.Duration, &c.MaxIono,
			&c.MinIono, &c.AvgIono, &c.MaxNdvi, &c.MinNdvi,
			&c.AvgNdvi, &c.MaxRad, &c.MinRad, &c.AvgRad, &c.MaxSpec,
			&c.MinSpec, &c.AvgSpec)
		if err != nil {
			return nil, errors.Wrap(err, "Error scanning rows")
		}
		computations = append(computations, c)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "Error scanning rows")
	}

	return computations, nil
}
