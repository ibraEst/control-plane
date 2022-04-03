package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterAndListGateways(t *testing.T) {

	gtwService := NewGatewayService([]Gateway{})
	server := NewControlPlane(gtwService, nil)
	s := SetupHttpServer(8080, server)

	gtw1 := Gateway{ID: "1", Name: "gtw-1", Version: 1, Region: "europe-west1"}
	gtw2 := Gateway{ID: "2", Name: "gtw-2", Version: 2, Region: "europe-west2"}

	s.Handler.ServeHTTP(httptest.NewRecorder(), newPostGatewayRequest(gtw1))
	s.Handler.ServeHTTP(httptest.NewRecorder(), newPostGatewayRequest(gtw2))

	t.Run("list registered gateways", func(t *testing.T) {

		w := httptest.NewRecorder()
		s.Handler.ServeHTTP(w, newGetGatewaysRequest())
		assertStatus(t, w.Code, http.StatusOK)
		got := getGatewaysFromResponse(t, w.Body)

		want := []Gateway{gtw1, gtw2}
		assertGateways(t, got, want)

	})

}

func newGetGatewaysRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/v1/gateways", nil)
	return req
}

func newPostGatewayRequest(gtw Gateway) *http.Request {
	data, _ := json.Marshal(gtw)
	req, _ := http.NewRequest(http.MethodPost, "/v1/gateways", bytes.NewBuffer(data))
	return req
}

func TestGetConfigurationByInstanceId(t *testing.T) {

	registeredConfigs := initConfigurations()
	configService := NewConfigurationServiceImpl(registeredConfigs)
	server := NewControlPlane(nil, configService)
	s := SetupHttpServer(8080, server)

	t.Run("get configurations by gateway id", func(t *testing.T) {

		resp := httptest.NewRecorder()
		s.Handler.ServeHTTP(resp, newGetConfigurationByIdRequest("2"))

		assertStatus(t, resp.Code, http.StatusOK)
		actual := getConfigurationFromResponse(t, resp.Body)
		assertConfigResponse(t, actual, registeredConfigs["2"])

	})
}

func newGetConfigurationByIdRequest(id string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, "/v1/gateways/"+id+"/configurations", nil)
	return req
}

func initConfigurations() map[string][]Configuration {
	configStore := map[string][]Configuration{}
	configStore["1"] = []Configuration{{
		Direction:  "spokeToHub",
		ServiceId:  "s1",
		Ports:      nil,
		ServiceVIP: "10.10.10.10",
		ServiceIP:  "2.2.2.2",
	}}

	configStore["2"] = []Configuration{{
		Direction:  "spokeToHub",
		ServiceId:  "s2",
		Ports:      nil,
		ServiceVIP: "10.10.10.11",
		ServiceIP:  "2.2.2.3",
	},
		{
			Direction:  "HubToSpoke",
			ServiceId:  "s3",
			Ports:      nil,
			ServiceVIP: "10.10.13.13",
			ServiceIP:  "2.2.4.4",
		}}
	return configStore
}
