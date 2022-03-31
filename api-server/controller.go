package main

import (
	"context"
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

// GatewayService is an interface for gateways management.
type GatewayService interface {
	GetGatewayInfo(name string) (Gateway, error)
	RegisterGateway(gtw Gateway) error
	GetGateways() []Gateway
}

// ConfigurationService is an interface for configuration management.
type ConfigurationService interface {
	GetConfiguration() ([]string, error)
}

func NewControlPlane(service GatewayService) *ControlPlane {

	cp := new(ControlPlane)
	cp.gatewayService = service

	return cp
}

func NewControlPlaneV2(configService ConfigurationService) *ControlPlane {

	cp := new(ControlPlane)
	cp.configurationService = configService

	return cp
}

func (cp *ControlPlane) registerGateway(c *gin.Context) {
	log.Println("Endpoint Hit: registerGateway")

	var gateway Gateway
	if err := c.BindJSON(&gateway); err != nil {
		return
	}

	err := cp.gatewayService.RegisterGateway(gateway)
	if err != nil {
		c.JSON(http.StatusBadRequest, gateway)
		return
	}

	c.JSON(http.StatusCreated, gateway)

	log.Println("New Gateway Successfully Created")

}

func (cp *ControlPlane) listGateways(c *gin.Context) {
	fmt.Println("Endpoint Hit: listGateways")

	result := cp.gatewayService.GetGateways()
	fmt.Println(result)
	c.JSON(http.StatusOK, result)
}

func (cp *ControlPlane) returnSingleGateway(c *gin.Context) {
	fmt.Println("Endpoint Hit: returnSingleGateway")

	key := c.Param("id")
	result, err := cp.gatewayService.GetGatewayInfo(key)

	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, result)

}

func (cp *ControlPlane) GetConfiguration(c *gin.Context) {
	fmt.Println("Endpoint Hit: get configuration")

	result, _ := cp.configurationService.GetConfiguration()
	//fmt.Println(result)
	c.JSON(http.StatusOK, result)
}

func (cp *ControlPlane) runHTTPServer(ctx context.Context, port int, controller *ControlPlane) error {

	router := gin.Default()

	router.POST("/v1/gateways", controller.registerGateway)
	router.GET("/v1/gateways", controller.listGateways)
	router.GET("/v1/gateways/:id", controller.returnSingleGateway)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	log.Print("HTTP Server Listening on: ", port)

	return srv.ListenAndServe()

}
