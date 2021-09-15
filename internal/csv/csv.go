package csv

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/Simek13/satelliteApp/internal/satellites"
)

const dateLayout = "01-02-2006 15:04"

func ReadCsvFromUrl(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ParseCsvData(rows [][]string) (map[string]satellites.Satellite, error) {

	sats := make(map[string]satellites.Satellite)

	for _, row := range rows[1:] {

		satId := row[0]
		if _, ok := sats[satId]; !ok {
			sat := satellites.BasicSatellite{
				Id:               satId,
				Timestamps:       make([]time.Time, 0),
				IonoIndexes:      make([]float64, 0),
				NdviIndexes:      make([]float64, 0),
				RadiationIndexes: make([]float64, 0),
			}
			switch satId {
			case "30J14":
				sat.SatelliteType = satellites.Ea
				sats[satId] = &satellites.EaSatellite{
					BasicSatellite: sat,
					Altitudes:      make([]float64, 0),
				}
			case "13A14", "6N14":
				sat.SatelliteType = satellites.Ss
				sats[satId] = &satellites.SsSatellite{
					BasicSatellite: sat,
					SeaSalinities:  make([]float64, 0),
				}
			case "8J14":
				sat.SatelliteType = satellites.Vc
				sats[satId] = &satellites.VcSatellite{
					BasicSatellite: sat,
					Vegetations:    make([]string, 0),
				}
			}
		}

		tm, err := time.Parse(dateLayout, row[1])
		if err != nil {
			return nil, err
		}

		satellite := sats[satId]

		satellite.GetSatellite().Timestamps = append(satellite.GetSatellite().Timestamps, tm)

		val, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return nil, err
		}
		satellite.GetSatellite().IonoIndexes = append(satellite.GetSatellite().IonoIndexes, val)

		val, err = strconv.ParseFloat(row[3], 64)
		if err != nil {
			return nil, err
		}
		satellite.GetSatellite().NdviIndexes = append(satellite.GetSatellite().NdviIndexes, val)

		val, err = strconv.ParseFloat(row[4], 64)
		if err != nil {
			return nil, err
		}
		satellite.GetSatellite().RadiationIndexes = append(satellite.GetSatellite().RadiationIndexes, val)

		switch sat := satellite.(type) {
		case *satellites.EaSatellite:
			val, err := strconv.ParseFloat(row[5], 64)
			if err != nil {
				return nil, err
			}
			sat.Altitudes = append(sat.Altitudes, val)
		case *satellites.SsSatellite:
			val, err := strconv.ParseFloat(row[5], 64)
			if err != nil {
				return nil, err
			}
			sat.SeaSalinities = append(sat.SeaSalinities, val)
		case *satellites.VcSatellite:
			sat.Vegetations = append(sat.Vegetations, row[5])
		}
	}

	return sats, nil
}
