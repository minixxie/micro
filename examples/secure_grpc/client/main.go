package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/dakalab/micro/example/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	serverName = "server"
	serverCert = "certs/server.crt"
	clientCert = "certs/client.crt"
	clientKey  = "certs/client.key"
	ca         = "certs/ca.crt"
)

func main() {
	/***********************************************************************************************
		Client 1: client connect to insecure grpc server
	***********************************************************************************************/
	conn, _ := grpc.Dial("localhost:9999", grpc.WithInsecure())
	var client = proto.NewGreeterClient(conn)

	var ctx, cancel = context.WithTimeout(
		context.Background(),
		1*time.Second,
	)
	defer cancel()

	var req = &proto.HelloRequest{Name: "Hyper from insecure client"}
	res, err := client.SayHello(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.GetMessage())

	/***********************************************************************************************
		Client 2: client connect to tls grpc server, note that our demo server's name is "server"
	***********************************************************************************************/
	creds, err := credentials.NewClientTLSFromFile(serverCert, serverName)
	if err != nil {
		log.Fatal(err)
	}

	// create a connection with the TLS credentials
	conn2, err := grpc.Dial("localhost:19999", grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}
	var client2 = proto.NewGreeterClient(conn2)

	req = &proto.HelloRequest{Name: "Hyper from tls client"}
	res, err = client2.SayHello(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.GetMessage())

	/***********************************************************************************************
		Client 3: client connect to mutual tls grpc server with certificate authority
	************************************************************************************************/
	// load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(clientCert, clientKey)
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

	// create the TLS credentials for transport
	creds2 := credentials.NewTLS(&tls.Config{
		ServerName:   serverName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	conn3, err := grpc.Dial("localhost:29999", grpc.WithTransportCredentials(creds2))
	if err != nil {
		log.Fatal(err)
	}
	var client3 = proto.NewGreeterClient(conn3)

	req = &proto.HelloRequest{Name: "Hyper from mutual tls client"}
	res, err = client3.SayHello(ctx, req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.GetMessage())
}
