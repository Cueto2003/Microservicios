package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"context"
	"time"
	"os"
	"os/signal"
	"syscall"

	"proyecto/metadataUser/internal/controller"
	httphandler "proyecto/metadataUser/internal/handler"
	"proyecto/metadataUser/internal/repository/memory"
	"proyecto/pkg/discovery/consul"
	registrypkg"proyecto/pkg/registry"
)

const defaultServiceName = "metadata-user"


func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "Puerto del microservicio de metadata de usuario")
	flag.Parse()
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	log.Printf("Starting %s service on port %d", serviceName, port)


	// ===== Consul =====
	// Dónde está Consul (local: localhost:8500 (si no lo corro con docker) o si es docker: consul:8500 o docker->host: host.docker.internal:8500)
	consulAddr := os.Getenv("CONSUL_HOST")
	if consulAddr == "" {
		consulAddr = "localhost:8500"
	}

	log.Printf("CONSUL_HOST = %s", consulAddr)

	reg, err := consul.NewRegistry(consulAddr)
	if err != nil {
		log.Fatalf("error creando registry de consul: %v", err)
	}

	ctx := context.Background()
	instanceID := registrypkg.GenerateInstanceID(serviceName)

	// Dirección que se **publica** en Consul para que otros clientes la usen.
	// En local: localhost
	// En Docker (misma red que Consul): el nombre del contenedor (p.ej. auth-server)
	host := os.Getenv("SERVICE_HOST")
	if host == "" {
		// Fallback sensato: en local funciona; en Docker deberías pasar SERVICE_HOST=auth-server
		host = "localhost"
	}

	hostPort := fmt.Sprintf("%s:%d", host, port)

	// Registro en Consul (nota: 0.0.0.0 NO es válido para Consul)
	if err := reg.Register(ctx, instanceID, serviceName, hostPort); err != nil {
		log.Fatalf("error registrando servicio en consul: %v", err)
	}
	log.Printf("%s registrado en Consul con ID=%s y address=%s", serviceName, instanceID, hostPort)

	// Health TTL (latido): marca el servicio como "passing" periódicamente
	go func() {
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()
		for range t.C {
			if err := reg.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Printf("Failed to report healthy state: %v", err)
			}
		}
	}()

	// Deregistro limpio al salir por señal (Ctrl c /SIGTERM) se borra en consul 
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		log.Println("Recibida señal, deregistrando servicio en Consul...")
		_ = reg.Deregister(context.Background(), instanceID, serviceName)
		os.Exit(0)
	}()

	//Crear todo 
	r := memory.New()
	c := metadataUser.New(r)
	h := httphandler.New(c)

	// endpoint
	
	// Rutas 
	mux := http.NewServeMux()
	mux.Handle("/MetadataUser", http.HandlerFunc(h.CreateMetadatUser)) // Es para Crear los usuarios 
	mux.Handle("/MetadataUser/Get", http.HandlerFunc(h.GetMetadatUser)) //Es para obtener los usuarios 

	//Servidor HTTP
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second, // tiempo máx. para leer la petición
		WriteTimeout: 15 * time.Second, // tiempo máx. para escribir la respuesta
		IdleTimeout:  60 * time.Second, // tiempo máx. de conexión inactiva
	}
	// Graceful shutdown adicional (opcional): cierre ordenado del server
	go func() {
		<-stop
		log.Println("Shutting down HTTP server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	log.Printf("%s escuchando en %s", serviceName, srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

/* Le faltan algunas cosas, como verificar que si tenga acceso el usuario , y la parte de modificar, y borrar */