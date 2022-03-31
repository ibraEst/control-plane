package main

import (
	"context"
	"log"
	"os"
	"strconv"
)

const defaultPort = 10000

func main() {

	cp := NewControlPlane(NewGatewayService())
	port, err := strconv.Atoi(os.Getenv("CONTROL_PLANE_PORT"))

	if err != nil {
		port = defaultPort
	}
	err2 := cp.runHTTPServer(context.Background(), port, cp)
	if err2 != nil {
		log.Fatal(err2)
	}

	log.Println("Control plane server started..")

}
