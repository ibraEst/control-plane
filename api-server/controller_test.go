package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

//Mock object which meet the gateway service interface requirements.
type MockGatewayService struct {
	records []Gateway
}

func (s *MockGatewayService) GetGatewayInfo(id string) (Gateway, error) {
	var result Gateway
	for _, gateway := range s.records {
		if gateway.ID == id {
			return gateway, nil
		}
	}
	return result, errors.New("No gateway registred with id " + id)
}

func (s *MockGatewayService) RegisterGateway(gtw Gateway) error {
	found, _ := s.GetGatewayInfo(gtw.ID)
	if (Gateway{} != found) { //replace with not nil func
		return errors.New("Gateway already registred with id " + gtw.ID)
	}

	s.records = append(s.records, gtw)
	return nil
}

func (s *MockGatewayService) GetGateways() []Gateway {
	return s.records
}

//Mock object which meet the configuration service interface requirements.
type MockConfigService struct {
	configurations map[string][]Configuration
}

func (s *MockConfigService) GetConfiguration(id string) ([]Configuration, error) {
	v, found := s.configurations[id]

	if !found {
		return nil, errors.New("Gateway with id " + id + " not registered")
	}

	return v, nil
}

func TestListGateways(t *testing.T) {

	gateways := []Gateway{
		{ID: "1",
			Name: "gtw-1", Version: 2, Region: "euw1",
		},
		{ID: "2",
			Name: "gtw-2", Version: 3, Region: "euw1",
		},
	}
	mockGatewayService := MockGatewayService{
		gateways,
	}

	server := NewControlPlane(&mockGatewayService, nil)
	port := 8080
	s := SetupHttpServer(port, server)

	// Create the mock request you'd like to test.
	req, err := http.NewRequest(http.MethodGet, "/v1/gateways", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// Create a response recorder so you can inspect the response
	w := httptest.NewRecorder()

	// Perform the request
	s.Handler.ServeHTTP(w, req)

	// Check to see if the response was what you expected
	assertStatus(t, w.Code, http.StatusOK)
	got := getGatewaysFromResponse(t, w.Body)

	assertGateways(t, got, gateways)

}

func TestGETGateways(t *testing.T) {

	gateways := []Gateway{
		{ID: "1",
			Name: "gtw-1", Version: 2, Region: "euw1",
		},
		{ID: "2",
			Name: "gtw-2", Version: 3, Region: "euw1",
		},
	}
	mockGatewayService := MockGatewayService{
		gateways,
	}

	server := NewControlPlane(&mockGatewayService, nil)
	port := 8080
	s := SetupHttpServer(port, server)

	t.Run("returns 1st gateway info", func(t *testing.T) {
		//given
		name := "1"
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/gateways/%s", name), nil)
		resp := httptest.NewRecorder()
		//when

		s.Handler.ServeHTTP(resp, req)
		//then
		assertStatus(t, resp.Code, http.StatusOK)

		actual := getOneGatewayFromResponse(t, resp.Body)

		assertGateway(t, actual, gateways[0])

	})

	t.Run("returns 404 in unknown gateway", func(t *testing.T) {
		//given
		req, _ := http.NewRequest(http.MethodGet, "/v1/gateways/unknown", nil)
		resp := httptest.NewRecorder()
		//when
		s.Handler.ServeHTTP(resp, req)

		//then
		assertStatus(t, resp.Code, http.StatusNotFound)

	})

}

func TestPostGateway(t *testing.T) {

	gateways := []Gateway{{ID: "1",
		Name: "gtw-1", Version: 2, Region: "euw1",
	}}
	mockGatewayService := MockGatewayService{
		gateways,
	}

	server := NewControlPlane(&mockGatewayService, nil)
	port := 8080
	s := SetupHttpServer(port, server)

	t.Run("returns status created when new gateway", func(t *testing.T) {
		//given
		newGtw := Gateway{ID: "2", Name: "gtw-2", Version: 2, Region: "europe-west1"}

		data, _ := json.Marshal(newGtw)
		request, _ := http.NewRequest(http.MethodPost, "/v1/gateways", bytes.NewBuffer(data))
		response := httptest.NewRecorder()
		//when
		s.Handler.ServeHTTP(response, request)

		//then
		assertStatus(t, response.Code, http.StatusCreated)

	})

	t.Run("returns status bad request when gateway already registred", func(t *testing.T) {
		//given
		newGtw := gateways[0]

		data, _ := json.Marshal(newGtw)
		request, _ := http.NewRequest(http.MethodPost, "/v1/gateways", bytes.NewBuffer(data))
		response := httptest.NewRecorder()
		//when
		s.Handler.ServeHTTP(response, request)

		//then
		assertStatus(t, response.Code, http.StatusBadRequest)

	})

}

func TestGETConfiguration(t *testing.T) {

	store := make(map[string][]Configuration)
	store["1"] = []Configuration{{
		Direction:  "spokeToHub",
		ServiceId:  "s1",
		Ports:      nil,
		ServiceVIP: "10.10.10.10",
		ServiceIP:  "2.2.2.2",
	}}

	port := 8080
	t.Run("returns last configuration when identified gateway", func(t *testing.T) {
		//given list of configuration and valid gateway id
		expected := store["1"]
		mockConfigService := &MockConfigService{
			store,
		}

		server := NewControlPlane(nil, mockConfigService)
		s := SetupHttpServer(port, server)

		//when pull config
		req, _ := http.NewRequest(http.MethodGet, "/v1/gateways/1/configuration", nil)
		resp := httptest.NewRecorder()
		s.Handler.ServeHTTP(resp, req)

		//then receive one config

		assertStatus(t, resp.Code, http.StatusOK)
		actual := getConfigurationFromResponse(t, resp.Body)
		assertConfigResponse(t, actual, expected)

	})

	t.Run("returns 500 internal server error if unknown gateway id", func(t *testing.T) {
		//given list of configuration and valid gateway id
		mockConfigService := &MockConfigService{
			store,
		}

		server := NewControlPlane(nil, mockConfigService)
		s := SetupHttpServer(port, server)

		//when pull config
		req, _ := http.NewRequest(http.MethodGet, "/v1/gateways/UNKNOWN/configuration", nil)
		resp := httptest.NewRecorder()
		s.Handler.ServeHTTP(resp, req)

		//then receive one config

		assertStatus(t, resp.Code, http.StatusInternalServerError)

	})
	/*	t.Run("returns 400 bad request if is not valid request JSON", func(t *testing.T) {
			//given
			name := "1"
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/gateways/%s", name), nil)

			resp := httptest.NewRecorder()
			//when

			r.ServeHTTP(resp, req)
			//then
			assertStatus(t, resp.Code, http.StatusOK)

			actual := getOneGatewayFromResponse(t, resp.Body)

			assertGateway(t, actual, want[0])
		}




		t.Run("returns 500 internal server error if no configuration found for the gateway id", func(t *testing.T) {
			//given
			name := "1"
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/v1/gateways/%s", name), nil)

			resp := httptest.NewRecorder()
			//when

			r.ServeHTTP(resp, req)
			//then
			assertStatus(t, resp.Code, http.StatusOK)

			actual := getOneGatewayFromResponse(t, resp.Body)

			assertGateway(t, actual, want[0])
		}*/
}
func assertStatus(t *testing.T, actual, expected int) {
	t.Helper()
	if actual != expected {
		t.Errorf("got %d, want %d", actual, expected)
	}
}

func getGatewaysFromResponse(t testing.TB, body io.Reader) (result []Gateway) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&result)

	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of gateway, '%v'", body, err)
	}

	return
}

func getConfigurationFromResponse(t testing.TB, body io.Reader) (result []Configuration) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&result)

	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of configuration, '%v'", body, err)
	}

	return
}

func getOneGatewayFromResponse(t testing.TB, body io.Reader) (result Gateway) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&result)

	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of gateway, '%v'", body, err)
	}

	return
}

func assertGateways(t testing.TB, got, want []Gateway) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertGateway(t testing.TB, got, want Gateway) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func assertConfigResponse(t testing.TB, got, want []Configuration) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
