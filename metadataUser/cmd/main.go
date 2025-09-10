package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"context"
	"time"

	"proyecto/metadataUser/internal/controller"
	httphandler "proyecto/metadataUser/internal/handler"
	"proyecto/metadataUser/internal/repository/memory"
	"proyecto/pkg/discovery/consul"
	"proyecto/pkg/registry"
)

const serviceName = "metadata-user"

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "Puerto del microservicio de metadata de usuario")
	flag.Parse()

	log.Printf("Starting %s service on port %d", serviceName, port)
	//La parte de consul 
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)
	// wiring: repo -> controller -> handler
	r := memory.New()
	c := metadataUser.New(r)
	h := httphandler.New(c)

	// endpoint

	http.Handle("/MetadataUser", http.HandlerFunc(h.CreateMetadatUser))

	http.Handle("/MetadataUser/Get", http.HandlerFunc(h.GetMetadatUser))

	// server
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
