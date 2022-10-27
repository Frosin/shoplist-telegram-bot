package iot

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type Server struct {
	storage IOTStorage
	port    string
}

func NewServer(storage IOTStorage, port string) *Server {
	return &Server{
		storage: storage,
	}
}

func (s *Server) hello(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()

	log.Println("server got values:", values)

	for key, v := range values {
		if len(v) == 0 {
			log.Println("invalid value")
			continue
		}

		s.storage.SaveValues(time.Now(), key, v[0])
	}
}

func (s *Server) StartServer() {
	router := mux.NewRouter()
	router.HandleFunc("/hello", s.hello).Methods("GET")

	srv := &http.Server{
		Addr:    ":" + s.port,
		Handler: router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := http.ListenAndServe(":"+s.port, nil); err != nil && err != http.ErrServerClosed {
			log.Println("listen: %s\n", err)
		}
	}()
	log.Print("IOT Server Started")

	<-done
	log.Print("IOT Server Stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
}
