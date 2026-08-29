package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	host    string
	port    int
	handler http.Handler
}

func New(host string, port int, mux http.Handler) *Server {
	return &Server{host: host, port: port, handler: mux}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.host+":"+strconv.Itoa(s.port), s.handler)

}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	svr := New("localhost", 8000, mux)
	fmt.Println("Server started: http://localhost:8000")
	err := svr.Start()
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
