package main

import (
	"flag"
	"log"
)

var port int

func init() {
	flag.IntVar(&port, "port", 9090, "Server port")
}

func main() {

	flag.Parse()
	cp := NewControlPlane(NewGatewayService([]Gateway{}), NewConfigurationServiceImpl([]Configuration{}))
	s := SetupHttpServer(port, cp)
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Control plane server started and listening to port %d", port)

}
