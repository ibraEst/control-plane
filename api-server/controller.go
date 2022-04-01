package main

import (
	"fmt"
	"log"
	"net/http"

	// provides the REST API controller framework
	"github.com/gin-gonic/gin"
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

type ControlPlane struct {
	gatewayService       GatewayService
	configurationService ConfigurationService
}

func NewControlPlane(gatewayService GatewayService, configurationService ConfigurationService) *ControlPlane {

	return &ControlPlane{gatewayService: gatewayService, configurationService: configurationService}
}

// GatewayService is an interface for gateways management.
type GatewayService interface {
	GetGatewayInfo(name string) (Gateway, error)
	RegisterGateway(gtw Gateway) error
	GetGateways() []Gateway
}

// ConfigurationService is an interface for configuration management.
type ConfigurationService interface {
	GetConfiguration(id string) ([]Configuration, error)
}

func (cp *ControlPlane) registerGateway(c *gin.Context) {
	log.Println("Endpoint Hit: registerGateway")

	var gateway Gateway
	if err := c.BindJSON(&gateway); err != nil {
		log.Printf("error while parsing gateway :%v", gateway)
		c.Status(http.StatusBadRequest)
		return
	}

	err := cp.gatewayService.RegisterGateway(gateway)
	if err != nil {
		log.Println("error while registring gateway %v", gateway)
		c.JSON(http.StatusBadRequest, gateway)
		return
	}

	c.JSON(http.StatusCreated, gateway)

	log.Println("New Gateway Successfully Created")

}

func (cp *ControlPlane) listGateways(c *gin.Context) {
	log.Println("Endpoint Hit: listGateways")

	result := cp.gatewayService.GetGateways()
	fmt.Println(result)
	c.JSON(http.StatusOK, result)
}

func (cp *ControlPlane) returnSingleGateway(c *gin.Context) {
	log.Println("Endpoint Hit: returnSingleGateway")

	key := c.Param("id")
	result, err := cp.gatewayService.GetGatewayInfo(key)

	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, result)

}

func (cp *ControlPlane) GetConfigurationByInstanceId(c *gin.Context) {
	log.Println("Endpoint Hit: get configuration")

	id := c.Param("id")
	result, err := cp.configurationService.GetConfiguration(id)
	if err != nil {
		c.Status(http.StatusInternalServerError)
	} else {
		c.JSON(http.StatusOK, result)
	}
}

func SetupHttpServer(port int, controller *ControlPlane) *http.Server {

	router := gin.Default()

	router.POST("/v1/gateways", controller.registerGateway)
	router.GET("/v1/gateways", controller.listGateways)
	router.GET("/v1/gateways/:id", controller.returnSingleGateway)
	router.GET("/v1/gateways/:id/configurations", controller.GetConfigurationByInstanceId)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	log.Println("HTTP Server Listening on: ", port)

	return srv

}
