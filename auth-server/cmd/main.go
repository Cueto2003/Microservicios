package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"proyecto/auth-server/internal/controller"
	"proyecto/auth-server/internal/handler"
	"proyecto/auth-server/internal/repository/memory"

	"proyecto/pkg/discovery/consul"
	registrypkg "proyecto/pkg/registry" // alias solo para usar registrypkg.GenerateInstanceID
)

const defaultServiceName = "auth-server"

func main() {
	// Config básica del microservicio Auth-server
	var port int
	flag.IntVar(&port, "port", 8082, "Puerto del microservicio de Autenticación (auth-server)")
	flag.Parse()
	//para poder cambiar el nombre del microservicio si es que subo mas de 1 
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	//log para saber que se inicializó bien 
	log.Printf("Starting %s on port %d", serviceName, port)

	// ===== Consul =====
	// Dónde está Consul (local: localhost:8500 (si no lo corro con docker) o si es docker: consul:8500 o docker->host: host.docker.internal:8500)
	consulAddr := os.Getenv("CONSUL_HOST")
	if consulAddr == "" {
		consulAddr = "localhost:8500"
	}
	//para saber donde se esta colocando el microservicio 
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

	//Crear lo nesesario 
	repo := memory.New()
	ctrl := controller.New(repo)
	h := handler.New(ctrl, reg) // handler necesita el registry para llamar a metadata-user

	// Rutas 
	mux := http.NewServeMux()
	mux.Handle("/Auth-Server/register", http.HandlerFunc(h.RegisterUser))
	mux.Handle("/Auth-Server/login", http.HandlerFunc(h.Login))

	//Servidor HTTP
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second, // tiempo máx. para leer la petición
		WriteTimeout: 15 * time.Second, // tiempo máx. para escribir la respuesta
		IdleTimeout:  60 * time.Second, // tiempo máx. de conexión inactiva (keep-alive)
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
