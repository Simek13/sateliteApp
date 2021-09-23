package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net"
	"strconv"

	"github.com/Simek13/satelliteApp/internal/database"
	"github.com/doug-martin/goqu/v9"
	"google.golang.org/grpc"

	pb "github.com/Simek13/satelliteApp/internal/satellite_communication"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	log "github.com/sirupsen/logrus"
)

var cfg struct {
	serverPort string

	dbType string
	dbUser string
	dbPass string
	dbHost string
	dbPort string
	dbName string
}

type satelliteCommunicationServer struct {
	pb.UnimplementedSatelliteCommunicationServer

	db *database.MySQLDatabase
}

func (s *satelliteCommunicationServer) GetMeasurements(ctx context.Context, filter *pb.SatelliteFilter) (*pb.MeasurementResponse, error) {
	ctxlog := log.WithFields(log.Fields{"event": "server.GetMeasurements"})
	measurements, err := s.db.GetMeasurements(filter.Name)
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal(err)
	}
	pbMeasurements := make([]*pb.Measurement, 0)
	for _, m := range measurements {
		pbMeasurements = append(pbMeasurements, m.Protobuf())
	}

	return &pb.MeasurementResponse{Measurements: pbMeasurements}, nil

}

func (s *satelliteCommunicationServer) GetComputations(ctx context.Context, filter *pb.SatelliteFilter) (*pb.ComputationResponse, error) {
	ctxlog := log.WithFields(log.Fields{"event": "server.GetComputations"})
	computations, err := s.db.GetComputations(filter.Name)
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal(err)
	}
	pbComputations := make([]*pb.Computation, 0)
	for _, c := range computations {
		pbComputations = append(pbComputations, c.Protobuf())
	}

	computationResponse := &pb.ComputationResponse{Computations: pbComputations}

	return computationResponse, nil
}

func validate() (err error) {
	if cfg.dbType != "mysql" && cfg.dbType != "postgres" && cfg.dbType != "sqlite3" && cfg.dbType != "sqlserver" {
		return errors.New("invalid db type")
	}
	if len(cfg.dbUser) < 4 || len(cfg.dbUser) > 100 {
		return errors.New("db user is not between 4 and 100 characters")
	}
	if len(cfg.dbPass) < 4 || len(cfg.dbPass) > 100 {
		return errors.New("db password is not between 4 and 100 characters")
	}
	if len(cfg.dbHost) < 4 || len(cfg.dbHost) > 100 {
		return errors.New("db host is not between 4 and 100 characters")
	}
	var port int
	port, err = strconv.Atoi(cfg.dbPort)
	if err != nil {
		port, err = net.LookupPort("tcp", cfg.dbPort)
		if err != nil {
			return errors.New("invalid db port")
		}
	}
	if port < 1024 || port > 65535 {
		return errors.New("db port is not between 1024 and 65535")
	}
	if len(cfg.dbName) < 4 || len(cfg.dbName) > 100 {
		return errors.New("db name is not between 4 and 100 characters")
	}

	return nil
}

func main() {
	ctxlog := log.WithFields(log.Fields{"event": "server"})
	flag.StringVar(&cfg.serverPort, "server_port", ":10000", "Port that server listens on")
	flag.StringVar(&cfg.dbType, "db_type", "mysql", "type of database")
	flag.StringVar(&cfg.dbUser, "db_user", "root", "user name for database")
	flag.StringVar(&cfg.dbPass, "db_pass", "emis", "user password for database")
	flag.StringVar(&cfg.dbHost, "db_host", "127.0.0.1", "host for database")
	flag.StringVar(&cfg.dbPort, "db_port", "3306", "port for database connection")
	flag.StringVar(&cfg.dbName, "db_name", "satellites", "name of database")

	err := validate()
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal(err)
	}

	dbBaseUrl := fmt.Sprintf("%s:%s@tcp(%s:%s)/", cfg.dbUser, cfg.dbPass, cfg.dbHost, cfg.dbPort)
	dbUrl := dbBaseUrl + cfg.dbName
	db, err := sql.Open(cfg.dbType, dbUrl)
	if err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error opening database")
	}
	gdb := goqu.New(cfg.dbType, db)
	mysqlDb := &database.MySQLDatabase{gdb}

	fmt.Println("Successfully Connected to MySQL database")
	defer db.Close()

	lis, err := net.Listen("tcp", cfg.serverPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSatelliteCommunicationServer(s, &satelliteCommunicationServer{db: mysqlDb})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
