package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/doug-martin/goqu/v9"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Simek13/satelliteApp/internal/database"
	pb "github.com/Simek13/satelliteApp/pkg"
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
	measurements, err := s.db.GetMeasurements(int(filter.GetSatId()))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Measurements for satellite: %q could not be found", filter.GetSatId())
	}
	pbMeasurements := make([]*pb.Measurement, 0)
	for _, m := range measurements {
		pbMeasurements = append(pbMeasurements, m.Protobuf())
	}

	return &pb.MeasurementResponse{Measurements: pbMeasurements}, nil

}

func (s *satelliteCommunicationServer) GetComputations(ctx context.Context, filter *pb.SatelliteFilter) (*pb.ComputationResponse, error) {
	computations, err := s.db.GetComputations(int(filter.GetSatId()))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Computations for satellite: %q could not be found. %s", filter.GetSatId(), errors.Unwrap(err))
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

func run() error {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := []grpc.DialOption{grpc.WithInsecure()}
	gwmux := runtime.NewServeMux()
	err := pb.RegisterSatelliteCommunicationHandlerFromEndpoint(ctx, gwmux, "localhost"+cfg.serverPort, opts)
	if err != nil {
		return err
	}
	return http.ListenAndServe(":9090", gwmux)
}

func main() {
	ctxlog := log.WithFields(log.Fields{"event": "server"})
	flag.StringVar(&cfg.serverPort, "grpc-server-port", ":8080", "gRPC server port")
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
	log.Printf("Serving gRPC on localhost%s", cfg.serverPort)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	if err := run(); err != nil {
		ctxlog.WithFields(log.Fields{"status": "failed", "error": err}).Fatal("Error serving")
	}

}
