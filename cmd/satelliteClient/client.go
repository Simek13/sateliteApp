package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Simek13/satelliteApp/internal/database"
	"google.golang.org/grpc"

	pb "github.com/Simek13/satelliteApp/internal/satellite_communication"
	log "github.com/sirupsen/logrus"

	"github.com/namsral/flag"
)

var cfg struct {
	serverAddr string
}

func printMeasurements(client pb.SatelliteCommunicationClient, sf *pb.SatelliteFilter) error {
	fmt.Printf("Getting satellite measurements for satellite: %s \n", sf.Name)
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
	fmt.Printf("Getting satellite computations for satellite: %s \n", sf.Name)
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

	err = printMeasurements(client, &pb.SatelliteFilter{Name: "30J14"})
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error printing measurements")
	}

	err = printComputations(client, &pb.SatelliteFilter{Name: "13A14"})
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error printing computations")
	}
}
