package main

import (
	"errors"
	"log"
)

// GatewayServiceImpl collects data about gateways in memory.
type GatewayServiceImpl struct {
	list []Gateway
}

func NewGatewayService(list []Gateway) *GatewayServiceImpl {
	return &GatewayServiceImpl{list: list}
}

func (im *GatewayServiceImpl) GetGatewayInfo(id string) (Gateway, error) {
	var result Gateway

	for _, gateway := range im.list {
		if gateway.ID == id {
			return gateway, nil
		}
	}
	return result, errors.New("No gateway registred with id " + id)
}

func (im *GatewayServiceImpl) RegisterGateway(gtw Gateway) error {
	found, _ := im.GetGatewayInfo(gtw.ID)
	if (Gateway{}) != found { //todo:replace with not nil func
		log.Printf("Gateway %q will not be created as it is already registred", gtw.ID)
		return errors.New("No gateway registred with id " + gtw.ID)
	}
	im.list = append(im.list, gtw)
	return nil
}

func (im *GatewayServiceImpl) GetGateways() []Gateway {
	return im.list

}
