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

	"github.com/gin-gonic/gin"
)

//Mock object which meet the service interface requirements.
type ServiceMock struct {
	records []Gateway
}

type MockConfigService struct {
	getConfigFunc func() ([]string, error)
	//configurations []Configuration
}

func (s *ServiceMock) GetGatewayInfo(id string) (Gateway, error) {
	var result Gateway
	for _, gateway := range s.records {
		if gateway.ID == id {
			return gateway, nil
		}
	}
	return result, errors.New("No gateway registred with id " + id)
}

func (s *ServiceMock) RegisterGateway(gtw Gateway) error {
	found, _ := s.GetGatewayInfo(gtw.ID)
	if (Gateway{} != found) { //replace with not nil func
		return errors.New("Gateway already registred with id " + gtw.ID)
	}

	s.records = append(s.records, gtw)
	return nil
}

func (s *ServiceMock) GetGateways() []Gateway {
	return s.records
}

func (s *MockConfigService) GetConfiguration() ([]string, error) {

	return s.getConfigFunc()
}

func TestListGateways(t *testing.T) {

	want := []Gateway{
		{ID: "1",
			Name: "gtw-1", Version: 2, Region: "euw1",
		},
		{ID: "2",
			Name: "gtw-2", Version: 3, Region: "euw1",
		},
	}
	serviceMock := ServiceMock{
		want,
	}

	server := NewControlPlane(&serviceMock)

	// Switch to test mode so you don't get such noisy output
	gin.SetMode(gin.TestMode)

	// Setup router, and register routes
	r := gin.Default()
	r.GET("/v1/gateways", server.listGateways)

	// Create the mock request you'd like to test.
	req, err := http.NewRequest(http.MethodGet, "/v1/gateways", nil)
	if err != nil {
		t.Fatalf("Couldn't create request: %v\n", err)
	}

	// Create a response recorder so you can inspect the response
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Check to see if the response was what you expected
	assertStatus(t, w.Code, http.StatusOK)
	got := getGatewaysFromResponse(t, w.Body)

	assertGateways(t, got, want)

}

func TestGETGateways(t *testing.T) {

	want := []Gateway{
		{ID: "1",
			Name: "gtw-1", Version: 2, Region: "euw1",
		},
		{ID: "2",
			Name: "gtw-2", Version: 3, Region: "euw1",
		},
	}
	serviceMock := ServiceMock{
		want,
	}

	server := NewControlPlane(&serviceMock)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/v1/gateways/:id", server.returnSingleGateway)

	t.Run("returns 1st gateway info", func(t *testing.T) {
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

	})

	t.Run("returns 404 in unknown gateway", func(t *testing.T) {
		//given
		req, _ := http.NewRequest(http.MethodGet, "/v1/gateways/unknown", nil)
		resp := httptest.NewRecorder()
		//when
		r.ServeHTTP(resp, req)

		//then
		assertStatus(t, resp.Code, http.StatusNotFound)

	})

}

func TestPostGateway(t *testing.T) {

	gateways := []Gateway{{ID: "1",
		Name: "gtw-1", Version: 2, Region: "euw1",
	}}
	serviceMock := ServiceMock{
		gateways,
	}

	server := NewControlPlane(&serviceMock)

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/v1/gateways", server.registerGateway)

	t.Run("returns status created when new gateway", func(t *testing.T) {
		//given
		newGtw := Gateway{ID: "2", Name: "gtw-2", Version: 2, Region: "europe-west1"}

		data, _ := json.Marshal(newGtw)
		request, _ := http.NewRequest(http.MethodPost, "/v1/gateways", bytes.NewBuffer(data))
		response := httptest.NewRecorder()
		//when
		r.ServeHTTP(response, request)

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
		r.ServeHTTP(response, request)

		//then
		assertStatus(t, response.Code, http.StatusBadRequest)

	})

}

func TestGETConfiguration(t *testing.T) {

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	//register routes

	// Query string parameters are parsed using the existing underlying request object.
	// The request responds to a url matching:  /welcome?firstname=Jane&lastname=Doe
	r.GET("/welcome", func(c *gin.Context) {
		firstname := c.DefaultQuery("firstname", "Guest")
		lastname := c.Query("lastname") // shortcut for c.Request.URL.Query().Get("lastname")

		c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
	})

	t.Run("return status OK when valid request", func(t *testing.T) {
		//given
		req, _ := http.NewRequest(http.MethodGet, "/welcome?firstname=ben&lastname=ibra", nil)
		resp := httptest.NewRecorder()

		//when
		r.ServeHTTP(resp, req)
		//then
		assertStatus(t, resp.Code, http.StatusOK)
		got := resp.Body.String()
		want := "Hello ben ibra"
		if got != want {
			t.Errorf("got %v but want %v", got, want)
		}
	})

	t.Run("returns last configuration when identified gateway", func(t *testing.T) {
		//given list of configuration
		expected := []string{"config1"}
		serviceMock := &MockConfigService{
			getConfigFunc: func() ([]string, error) {
				return expected, nil
			},
		}

		server := NewControlPlaneV2(serviceMock)
		r.GET("/v1/gateways/configuration", server.GetConfiguration)

		//when pull config
		req, _ := http.NewRequest(http.MethodGet, "/v1/gateways/configuration", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		//then receive one config

		assertStatus(t, resp.Code, http.StatusOK)
		actual := getConfigurationFromResponse(t, resp.Body)
		assertConfigResponse(t, actual, expected)

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

		t.Run("returns 500 internal server error if unknown gateway id", func(t *testing.T) {
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

func getConfigurationFromResponse(t testing.TB, body io.Reader) (result []string) {
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

func assertConfigResponse(t testing.TB, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}
