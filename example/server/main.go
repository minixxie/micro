package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/dakalab/micro"
	"github.com/dakalab/micro/example/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Greeter - implementation of GreeterServer
type Greeter struct {
}

// SayHello implements gRPC endpoint "SayHello"
func (s *Greeter) SayHello(
	ctx context.Context,
	req *proto.HelloRequest,
) (*proto.HelloResponse, error) {
	return &proto.HelloResponse{
		Message: "Hello " + req.Name,
	}, nil
}

var _ proto.GreeterServer = (*Greeter)(nil) // make sure it implements the interface

var (
	serverName = "server"
	crt        = "certs/server.crt"
	key        = "certs/server.key"
	ca         = "certs/ca.crt"
)

func main() {

	reverseProxyFunc := func(
		ctx context.Context,
		mux *runtime.ServeMux,
		grpcHostAndPort string,
		opts []grpc.DialOption,
	) error {
		return proto.RegisterGreeterHandlerFromEndpoint(ctx, mux, grpcHostAndPort, opts)
	}

	// add swagger definition endpoint
	route := micro.Route{
		Method:  "GET",
		Pattern: micro.PathPattern("hello.swagger.json"),
		Handler: func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			data, _ := ioutil.ReadFile("proto/hello.swagger.json")
			w.Write(data)
		},
	}

	sf := func() {
		log.Println("Server shutting down")
	}

	// init redoc, enable api docs on http://localhost:8888
	redoc := &micro.RedocOpts{
		Up: true,
	}
	redoc.AddSpec("Greeter", "/hello.swagger.json")

	/***********************************************************************************************
		Server 1: insecure server
	***********************************************************************************************/
	s := micro.NewService(
		micro.Debug(true),
		micro.RouteOpt(route),
		micro.ShutdownFunc(sf),
		micro.Redoc(redoc),
	)
	proto.RegisterGreeterServer(s.GRPCServer, &Greeter{})

	/***********************************************************************************************
		Server 2: tls server with server-side encryption that does not expect client authentication
		or credentials
	************************************************************************************************/
	// create the TLS credentials
	serverCreds, err := credentials.NewServerTLSFromFile(crt, key)
	if err != nil {
		log.Fatal(err)
	}

	clientCreds, err := credentials.NewClientTLSFromFile(crt, serverName)
	if err != nil {
		log.Fatal(err)
	}

	s2 := micro.NewService(
		micro.Debug(true),
		micro.RouteOpt(route),
		micro.ShutdownFunc(sf),
		micro.Redoc(redoc),
		micro.GRPCServerOption(grpc.Creds(serverCreds)),
		micro.GRPCDialOption(grpc.WithTransportCredentials(clientCreds)),
	)
	proto.RegisterGreeterServer(s2.GRPCServer, &Greeter{})

	/***********************************************************************************************
		Server 3: mutual tls server with certificate authority
	************************************************************************************************/
	// load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(crt, key)
	if err != nil {
		log.Fatal(err)
	}

	// create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(ca)
	if err != nil {
		log.Fatal(err)
	}

	// append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append ca certs")
	}

	// create the TLS configuration
	serverCreds2 := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	clientCreds2 := credentials.NewTLS(&tls.Config{
		ServerName:   serverName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	s3 := micro.NewService(
		micro.Debug(true),
		micro.RouteOpt(route),
		micro.ShutdownFunc(sf),
		micro.Redoc(redoc),
		micro.GRPCServerOption(grpc.Creds(serverCreds2)),
		micro.GRPCDialOption(grpc.WithTransportCredentials(clientCreds2)),
	)
	proto.RegisterGreeterServer(s3.GRPCServer, &Greeter{})

	errChan := make(chan error, 1)

	// run insecure server 1
	go func() {
		var httpPort, grpcPort uint16
		httpPort = 8888
		grpcPort = 9999
		errChan <- s.Start(httpPort, grpcPort, reverseProxyFunc)
	}()

	// run tls server 2
	go func() {
		var httpPort, grpcPort uint16
		httpPort = 18888
		grpcPort = 19999
		errChan <- s2.Start(httpPort, grpcPort, reverseProxyFunc)
	}()

	// run mutual tls server 3
	go func() {
		var httpPort, grpcPort uint16
		httpPort = 28888
		grpcPort = 29999
		errChan <- s3.Start(httpPort, grpcPort, reverseProxyFunc)
	}()

	log.Fatal(<-errChan)
}
