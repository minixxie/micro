package main

import (
	"database/sql"
	proto_golang_grpc_v1 "golang_grpc/proto/golang_grpc/v1"
	"log"

	"github.com/minixxie/micro"
	"google.golang.org/grpc"
	_ "gopkg.in/go-sql-driver/mysql.v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/context"

	"golang_grpc/internal/lib"
	"golang_grpc/internal/models"
	"golang_grpc/internal/services"
)

func main() {

	//###########################//
	// LoadConfig
	//###########################//
	config := models.Config{}
	err := lib.LoadConfig(&config)
	if err != nil {
		log.Fatalf("Cannot load config file:", err)
	}

	//###########################//
	// MariaDB: mainDB
	//###########################//
	log.Printf("DB to connect: %s", config.Dbs.Main.Uri)
	mainDB, openDBErr := sql.Open("mysql", config.Dbs.Main.Uri)
	if openDBErr != nil {
		log.Fatal("Cannot open DB:", openDBErr)
	}
	defer mainDB.Close()
	pingDBErr := mainDB.Ping()
	if pingDBErr != nil {
		log.Fatal("Cannot ping DB:", pingDBErr)
	}
	mainDB.SetMaxIdleConns(config.Dbs.Main.MaxIdleConns)
	mainDB.SetMaxOpenConns(config.Dbs.Main.MaxOpenConns)
	// Validate DB schema version
	schemaVersion, ok := lib.CheckMySQLSchemaVersion(mainDB, config.Dbs.Main.ExpectedSchemaVersion)
	if !ok {
		log.Fatalf("Incorrect DB schema version: %d, expected: %d", schemaVersion, config.Dbs.Main.ExpectedSchemaVersion)
	}

	redoc := &micro.RedocOpts{
		Up: true,
	}
	s := micro.NewService(
		micro.Redoc(redoc),
	)

	// FirstService
	proto_golang_grpc_v1.RegisterFirstServiceServer(s.GRPCServer, &services.FirstService{
		FirstModel: &models.FirstModelImpl{
			MainDB: mainDB,
		},
	})
	// SecondService
	proto_golang_grpc_v1.RegisterSecondServiceServer(s.GRPCServer, &services.SecondService{})

	err = s.Start(80, 8080, func(ctx context.Context, mux *runtime.ServeMux, grpcHostAndPort string, opts []grpc.DialOption) error {
		var err error
		// FirstService
		err = proto_golang_grpc_v1.RegisterFirstServiceHandlerFromEndpoint(
			ctx, mux, grpcHostAndPort, opts)
		if err != nil {
			return err
		}
		// SecondService
		err = proto_golang_grpc_v1.RegisterSecondServiceHandlerFromEndpoint(
			ctx, mux, grpcHostAndPort, opts)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Printf("Cannot start service: %v", err)
		return
	}
}
