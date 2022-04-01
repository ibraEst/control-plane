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

	//Package metadata provides access to Google Compute Engine (GCE) metadata and API service accounts.
	"cloud.google.com/go/compute/metadata"
)

type Gateway struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version int    `json:"version"`
	Region  string `json:"region"`
}

type Configuration struct {
	Direction  string `json:"direction"`
	ServiceId  string `json:"serviceId"`
	Ports      []int  `json:"ports"`
	ServiceVIP string `json:"vip"`
	ServiceIP  string `json:"ip"`
}

const (
	endpointBasePath  = "/v1/gateways"
	metdataPathFormat = "/instance/service-accounts/default/identity?audience=%s"
)

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
	/*tokenURL := fmt.Sprintf(metdataPathFormat, serviceURL)
	idToken, err := metadata.Get(tokenURL)
	if err != nil {
		log.Fatalf("metadata.Get: failed to query id_token: %+v", err)
		return
	}
	*/
	idToken := ""
	//identity := Default()
	identity := &Gateway{
		ID:      "1",
		Name:    "gtw-1",
		Version: 2,
		Region:  "euw1",
	}

	//register once when starting
	registerGateway(serviceURL, identity, idToken)

	//TODO (NCP-384) ping server to get last configuration
	interval := 5 //to be configured
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for ; true; <-ticker.C {
		log.Println("Calling server to pull my configuration")
		pullConfiguration(serviceURL, identity.ID, idToken)
		log.Println("config pulled")
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

func pullConfiguration(url string, id string, idToken string) {

	req, err := http.NewRequest("GET", url+endpointBasePath+"/"+id+"/configuration", nil)
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
	log.Println("parsing response...", responseData)
	var responseObject []Configuration
	err = json.Unmarshal(responseData, &responseObject)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("response parsed")
	fmt.Println(responseObject)
}
