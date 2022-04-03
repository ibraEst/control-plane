package main

import (
	"errors"
)

// GatewayServiceImpl collects data about gateways in memory.
type ConfigurationServiceImpl struct {
	configStore map[string][]Configuration
}

func NewConfigurationServiceImpl(configStore map[string][]Configuration) *ConfigurationServiceImpl {
	return &ConfigurationServiceImpl{configStore: configStore}
}

func (c ConfigurationServiceImpl) GetConfiguration(id string) ([]Configuration, error) {
	v, found := c.configStore[id]

	if !found {
		return nil, errors.New("No configuration found for gateway id " + id)
	}
	return v, nil
}

func (c ConfigurationServiceImpl) GetConfigurations() map[string][]Configuration {
	return c.configStore
}
