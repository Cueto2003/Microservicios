package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"context"
	"time"

	"proyecto/auth-server/internal/controller"
    "proyecto/auth-server/internal/handler"
    "proyecto/auth-server/internal/repository/memory"
	"proyecto/pkg/discovery/consul"
	"proyecto/pkg/registry"
)

const serviceName = "auth-server"

func main() {
	var port int
	flag.IntVar(&port, "port", 8082, "Puerto del microservicio de Autenticación (auth-server)")
	flag.Parse()

	log.Printf("Starting %s service on port %d", serviceName, port)
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
	c := controller.New(r)
	h := handler.New(c)

	// endpoint
	// Registarar un usuario , esto implica tambien llamar al otro microservicio para crear la otra parte del usuario.
	
	http.Handle("/Auth-Server/register", http.HandlerFunc(h.RegisterUser))
	//valida el email + contraseña, devuelve un JWT o token similar.
	http.Handle("/Auth-Server/login", http.HandlerFunc(h.Login))
	//Verifica si el token esta todavia activo o si ya no 
	//http.Handle("/Auth-Server/VerifyToken", http.HandlerFunc(h.VerifyToken))


	// server
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}