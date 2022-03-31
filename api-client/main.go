package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/compute/metadata"
)

type Gateway struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version int    `json:"version"`
	Region  string `json:"region"`
}

const endpointBasePath = "/v1/gateways"

func Default() *Gateway {
	instanceID, _ := metadata.InstanceID()
	instanceName, _ := metadata.InstanceName()
	zone, _ := metadata.Zone()
	return &Gateway{ID: instanceID, Name: instanceName, Version: 0, Region: zone[:len(zone)-2]}
}

func main() {

	log.Println("Agent gateway started")

	//get control plane service URI
	serviceURL := os.Getenv("CONTROL_PLANE_URL")
	if serviceURL == "" {
		log.Fatalln("CONTROL_PLANE_URL environment variable is not provided")
		return
	}

	//get access token
	tokenURL := fmt.Sprintf("/instance/service-accounts/default/identity?audience=%s", serviceURL)
	idToken, err := metadata.Get(tokenURL)
	if err != nil {
		log.Fatalf("metadata.Get: failed to query id_token: %+v", err)
		return
	}

	identity := Default()

	//register once when starting
	registerGateway(serviceURL, identity, idToken)

	//TODO (NCP-384) ping server to get last configuration
	interval := 5 //to be configured
	ticker := time.NewTicker(time.Duration(interval) * time.Minute)
	for ; true; <-ticker.C {
		log.Println("Calling server to get my configuration")
	}

}

func registerGateway(url string, body *Gateway, idToken string) {

	var jsonStr, _ = json.Marshal(&body)

	req, err := http.NewRequest("POST", url+endpointBasePath, bytes.NewBuffer(jsonStr))
	if err != nil {
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", idToken))
	response, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var responseObject Gateway
	err2 := json.Unmarshal(responseData, &responseObject)
	if err2 != nil {
		log.Fatal(err2)
	}

	fmt.Println(responseObject)
}
