package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Simek13/satelliteApp/internal/database"
	"google.golang.org/grpc"

	pb "github.com/Simek13/satelliteApp/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/namsral/flag"
)

var cfg struct {
	serverAddr string
}

func printMeasurements(client pb.SatelliteCommunicationClient, sf *pb.SatelliteFilter) error {
	fmt.Printf("Getting satellite measurements for satellite: %v \n", sf.GetSatId())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	measurementResponse, err := client.GetMeasurements(ctx, sf)
	if err != nil {
		return err
	}
	pbMeasurements := measurementResponse.Measurements
	measurements := make([]database.Measurement, 0)
	for _, m := range pbMeasurements {
		measurements = append(measurements, *database.NewMeasurement(m))
	}
	for _, m := range measurements {
		fmt.Println(m)
	}
	return nil
}

func printComputations(client pb.SatelliteCommunicationClient, sf *pb.SatelliteFilter) error {
	fmt.Printf("Getting satellite computations for satellite: %v \n", sf.GetSatId())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	computationResponse, err := client.GetComputations(ctx, sf)
	if err != nil {
		return err
	}
	pbComputations := computationResponse.Computations
	computations := make([]database.Computation, 0)
	for _, c := range pbComputations {
		computations = append(computations, *database.NewComputation(c))
	}
	for _, c := range computations {
		fmt.Println(c)
	}

	return nil
}

func addSatellite(client pb.SatelliteCommunicationClient, s *pb.Satellite) error {
	fmt.Printf("Adding new satellite: %v \n", s)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.AddSatellite(ctx, s)
	if err != nil {
		return err
	}
	return nil
}

func addMeasurement(client pb.SatelliteCommunicationClient, m *pb.Measurement) error {
	fmt.Printf("Adding new measurement: %v \n", m)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.AddMeasurement(ctx, m)
	if err != nil {
		return err
	}
	return nil
}

func addComputation(client pb.SatelliteCommunicationClient, c *pb.Computation) error {
	fmt.Printf("Adding new computation: %v \n", c)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.AddComputation(ctx, c)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	ctxlog := log.WithFields(log.Fields{"event": "client"})
	flag.StringVar(&cfg.serverAddr, "server address", "localhost:10000", "Address of server")

	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial(cfg.serverAddr, opts...)
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Cannot connect to server.")
	}
	defer conn.Close()
	client := pb.NewSatelliteCommunicationClient(conn)

	/* err = printMeasurements(client, &pb.SatelliteFilter{SatId: 1})
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error printing measurements")
	}

	err = printComputations(client, &pb.SatelliteFilter{SatId: 2})
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error printing computations")
	} */

	err = addSatellite(client, &pb.Satellite{Name: "AG12"})
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error adding satellite")
	}

	err = addMeasurement(client, &pb.Measurement{FileName: "novi.csv", IdSat: 5, Timestamp: "15:15", IonoIndex: 2, NdviIndex: 5, RadiationIndex: 26, SpecificMeasurement: "china"})
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error adding measurement")
	}

	err = addComputation(client, &pb.Computation{IdSat: 5, Duration: "30:00", MaxIono: 5, MinIono: 2, AvgIono: 3, MaxNdvi: 15, MinNdvi: 12, AvgNdvi: 13, MaxRad: 3, MinRad: 1, AvgRad: 2, MaxSpec: 0, MinSpec: 0, AvgSpec: 0})
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error adding computation")
	}
}
