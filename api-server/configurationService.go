package main

// GatewayServiceImpl collects data about gateways in memory.
type ConfigurationServiceImpl struct {
	list []Configuration
}

func (c ConfigurationServiceImpl) GetConfiguration(id string) ([]Configuration, error) {
	//TODO implement me
	panic("implement me")
}

func NewConfigurationServiceImpl(list []Configuration) *ConfigurationServiceImpl {
	return &ConfigurationServiceImpl{list: list}
}
