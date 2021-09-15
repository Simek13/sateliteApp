package csv

import (
	"encoding/csv"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Simek13/satelliteApp/internal/satellites"
)

func TestParseCsvData(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
		want     map[string]satellites.Satellite
		wantErr  bool
	}{
		{"happypath",
			"fixtures/happypath.csv",
			map[string]satellites.Satellite{
				"30J14": &satellites.EaSatellite{
					BasicSatellite: satellites.BasicSatellite{
						Id:               "30J14",
						Timestamps:       []time.Time{time.Date(2016, 02, 20, 15, 19, 0, 0, time.UTC), time.Date(2016, 02, 20, 15, 21, 0, 0, time.UTC)},
						IonoIndexes:      []float64{5, 7},
						NdviIndexes:      []float64{29, 33},
						RadiationIndexes: []float64{32, 32.4},
						SatelliteType:    satellites.Ea,
					},
					Altitudes: []float64{830.9, 833.3},
				},
				"8J14": &satellites.VcSatellite{
					BasicSatellite: satellites.BasicSatellite{
						Id:               "8J14",
						Timestamps:       []time.Time{time.Date(2016, 02, 20, 15, 34, 0, 0, time.UTC)},
						IonoIndexes:      []float64{10},
						NdviIndexes:      []float64{49},
						RadiationIndexes: []float64{41},
						SatelliteType:    satellites.Vc,
					},
					Vegetations: []string{"WOODS"},
				},
				"6N14": &satellites.SsSatellite{
					BasicSatellite: satellites.BasicSatellite{
						Id:               "6N14",
						Timestamps:       []time.Time{time.Date(2016, 02, 20, 16, 04, 0, 0, time.UTC), time.Date(2016, 02, 20, 16, 06, 0, 0, time.UTC)},
						IonoIndexes:      []float64{19, 20},
						NdviIndexes:      []float64{54, 55},
						RadiationIndexes: []float64{47.6, 48.6},
						SatelliteType:    satellites.Ss,
					},
					SeaSalinities: []float64{2.2, 2.2},
				},
			},
			false},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.Open(tt.filepath)
			if err != nil {
				t.Fatalf("Error opening filepath, %v", err)
			}
			defer f.Close()

			csvReader := csv.NewReader(f)
			csvReader.Comma = ';'
			rows, err := csvReader.ReadAll()
			if err != nil {
				t.Fatalf("Error reading csv, %v", err)
			}
			got, err := ParseCsvData(rows)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCsvData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseCsvData() = %v, want %v", got, tt.want)
			}
		})
	}
}
