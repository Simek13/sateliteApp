package app

import (
	"errors"
	"fmt"

	"github.com/Simek13/satelliteApp/internal/database"
	"github.com/Simek13/satelliteApp/internal/print"
	"github.com/Simek13/satelliteApp/internal/satellites"

	log "github.com/sirupsen/logrus"
)

func Run(filename string, sats map[string]satellites.Satellite, mysqlDb *database.MySQLDatabase) {
	ctxlog := log.WithFields(log.Fields{"event": "main_loop"})

	print.PrintSatelliteMeasurementTimes(sats)

	fmt.Println()

	ionoAvg := make(map[string]float64)
	NDVIAvg := make(map[string]float64)
	radAvg := make(map[string]float64)
	altAvg := make(map[string]float64)
	salAvg := make(map[string]float64)

	// kako ovo raščlaniti na funkcije koje rade samo jednu stvar
	for id, sat := range sats {

		fmt.Println("Satellite: ", id)
		ionoCalc, ndviCalc, radCalc, specCalc := sat.Compute()
		fmt.Println("Ionosphere index:", ionoCalc[0], "(MIN)", ionoCalc[1], "(MAX)", ionoCalc[2], "(AVG)")
		ionoAvg[id] = ionoCalc[2]
		fmt.Println("NDVI index:", ndviCalc[0], "(MIN)", ndviCalc[1], "(MAX)", ndviCalc[2], "(AVG)")
		NDVIAvg[id] = ndviCalc[2]
		fmt.Println("Radiation index:", radCalc[0], "(MIN)", radCalc[1], "(MAX)", radCalc[2], "(AVG)")
		radAvg[id] = radCalc[2]
		switch sat.(type) {
		case *satellites.EaSatellite:
			fmt.Println("Earth altitude:", specCalc[0], "(MIN)", specCalc[1], "(MAX)", specCalc[2], "(AVG)")
			altAvg[id] = specCalc[2]
		case *satellites.SsSatellite:
			fmt.Println("Sea salinity index:", specCalc[0], "(MIN)", specCalc[1], "(MAX)", specCalc[2], "(AVG)")
			salAvg[id] = specCalc[2]
		}

		fmt.Println()

	}

	print.PrintSatelliteCalculationAverages(ionoAvg, NDVIAvg, radAvg, altAvg, salAvg)

	err := mysqlDb.AddSatellites(sats)
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal(errors.Unwrap(err))
	} else {
		ctxlog.WithFields(log.Fields{"status": "success", "event": "Successfully written satellites to db."}).Info()
	}

	err = mysqlDb.AddMeasurements(filename, sats)
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal(errors.Unwrap(err))
	} else {
		ctxlog.WithFields(log.Fields{"status": "success", "event": "Successfully written measurements to db."}).Info()
	}

	err = mysqlDb.AddComputations(sats)
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal(errors.Unwrap(err))
	} else {
		ctxlog.WithFields(log.Fields{"status": "success", "event": "Successfully written computations to db."}).Info()
	}
}
