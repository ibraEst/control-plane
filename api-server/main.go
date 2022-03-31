package main

import (
	"log"
	"os"
	"strconv"
)

const defaultPort = 10000

func main() {

	cp := NewControlPlane(NewGatewayService([]Gateway{}), NewConfigurationServiceImpl([]Configuration{}))
	port, err := strconv.Atoi(os.Getenv("CONTROL_PLANE_PORT"))

	if err != nil {
		port = defaultPort
	}
	s := SetupHttpServer(port, cp)
	err = s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Control plane server started..")

}
