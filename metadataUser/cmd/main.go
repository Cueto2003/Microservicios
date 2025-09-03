package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"proyecto/metadataUser/internal/controller"
	httphandler "proyecto/metadataUser/internal/handler"
	"proyecto/metadataUser/internal/repository/memory"
)

const serviceName = "metadata-user"

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "Puerto del microservicio de metadata de usuario")
	flag.Parse()

	log.Printf("Starting %s service on port %d", serviceName, port)

	// wiring: repo -> controller -> handler
	r := memory.New()
	c := metadataUser.New(r)
	h := httphandler.New(c)

	// endpoint
	http.Handle("/MetadataUser", http.HandlerFunc(h.CreateMetadatUser))

	// server
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
