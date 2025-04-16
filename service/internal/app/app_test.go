package app

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"

	config "github.com/OliveTin/OliveTin/internal/config"
)

func TestOliveTinApp(t *testing.T) {
	originalDir, err := os.Getwd()
	assert.Equal(t, err, nil)
	os.Chdir("../../") // go test sets the cwd to "app" by default
	defer os.Chdir(originalDir)

	type testCase struct {
		cfg  func() *config.Config
		test func(*testing.T)
	}

	testCases := map[string]testCase{
		"Test Page Loads": {
			cfg: func() *config.Config {
				return config.DefaultConfig()
			},
			test: func(t *testing.T) {
				url := "http://localhost:1337"
				expectedText := "<title>OliveTin</title>"
				failText := "Response did not contain the title"
				urlContainsText(t, url, expectedText, failText)

				url = "http://127.0.0.1:1337/webUiSettings.json"
				expectedText = "{\"BaseURL\":\".\",\"Rest\":\"./api/"
				failText = "UI Settings did not respond with the path"
				urlContainsText(t, url, expectedText, failText)
			},
		},
		"Test Page Loads on Subpath": {
			cfg: func() *config.Config {
				defaultConfig := config.DefaultConfig()
				defaultConfig.Subpath = "/subpath"
				return defaultConfig
			},
			test: func(t *testing.T) {
				url := "http://localhost:1337/subpath/"
				expectedText := "<title>OliveTin</title>"
				failText := "Response did not contain the title"
				urlContainsText(t, url, expectedText, failText)

				url = "http://127.0.0.1:1337/subpath/webUiSettings.json"
				expectedText = "{\"BaseURL\":\"/subpath\",\"Rest\":\"/subpath/api/"
				failText = "UI Settings did not respond with the path"
				urlContainsText(t, url, expectedText, failText)
			},
		},
		"Test that prometheus metrics are available": {
			cfg: func() *config.Config {
				defaultConfig := config.DefaultConfig()
				defaultConfig.Prometheus = config.PrometheusConfig{
					Enabled:          true,
					DefaultGoMetrics: true,
				}
				return defaultConfig
			},
			test: func(t *testing.T) {
				url := "http://127.0.0.1:1341"
				expectedText := "go_gc_duration_seconds"
				failText := "failed to scrape default go metrics"
				urlContainsText(t, url, expectedText, failText)

				expectedText = "olivetin_actions_requested_count"
				failText = "olivetin metrics not available"
				urlContainsText(t, url, expectedText, failText)
			},
		},
		"Test that the grpc API is available": {
			cfg: func() *config.Config {
				defaultConfig := config.DefaultConfig()
				return defaultConfig
			},
			test: func(t *testing.T) {
				url := "http://localhost:1337"
				expectedText := "<title>OliveTin</title>"
				failText := "Response did not contain the title"
				urlContainsText(t, url, expectedText, failText)

				url = "http://127.0.0.1:1337/api/GetDashboardComponents"
				expectedText = "authenticatedUser"
				failText = "GetDashboardComponents"
				urlContainsText(t, url, expectedText, failText)
			},
		},
		"Test that websockets and actions can be used": {
			cfg: func() *config.Config {
				defaultConfig := config.DefaultConfig()
				defaultConfig.Actions = append(defaultConfig.Actions, &config.Action{
					ID:            "ping",
					Title:         "ping",
					Shell:         "ping example.com -c 1",
					Icon:          "ping",
					MaxConcurrent: 1,
					Timeout:       600,
				})
				return defaultConfig
			},
			test: func(t *testing.T) {
				url := "http://localhost:1337"
				expectedText := "<title>OliveTin</title>"
				failText := "Response did not contain the title"
				urlContainsText(t, url, expectedText, failText)

				url = "http://127.0.0.1:1337/api/GetDashboardComponents"
				expectedText = "\"title\":\"ping\""
				failText = "GetDashboardComponents"
				urlContainsText(t, url, expectedText, failText)

				// Dial the websocket
				client, err := websocketDialer(t, "ws://127.0.0.1:1337/websocket")
				defer client.Close()

				// Send a message over the websocket to start monitoring
				message := []byte("monitor")
				err = client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					t.Fatalf("could not write message to WebSocket server: %v", err)
				}

				// Start a "ping" action
				url = "http://127.0.0.1:1337/api/StartAction"
				body := "{\"actionId\":\"ping\",\"arguments\":[],\"uniqueTrackingId\":\"89c45ec7-8ed1-4cb2-8198-826286396dde\"}"
				expectedText = "executionTrackingId"
				failText = "failed to start an action"
				response := postContainsText(t, url, body, expectedText, failText)
				fmt.Println(response)

				// Read a message off the websocket
				_, websocketResponse, err := client.ReadMessage()
				if err != nil {
					t.Fatalf("could not read message from WebSocket server: %v", err)
				}
				websocketResponseString := string(websocketResponse)
				fmt.Println(websocketResponseString)

				assert.Contains(t, websocketResponseString, "olivetin.api.v1.EventExecutionStarted")
			},
		},
	}
	for testCaseName, testCase := range testCases {
		fmt.Printf("Starting the test: %s", testCaseName)
		cfg := testCase.cfg()
		s := CreateOliveTin(cfg)
		go s.Start()
		// Sleep to yield to ensure the ensure the test server starts
		// This is needed because Start() creates a lot of goroutines
		time.Sleep(100 * time.Millisecond)
		testCase.test(t)
		s.Stop()
	}
}

func websocketDialer(t *testing.T, url string) (*websocket.Conn, error) {
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("could not connect to WebSocket server: %v", err)
	}
	return client, err
}

func urlContainsText(t *testing.T, queryUrl, expectedText, failText string) {
	parsedUrl, _ := url.Parse(queryUrl)

	resp, err := http.Get(parsedUrl.String())
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)

	body, err := io.ReadAll(resp.Body)
	assert.Equal(t, err, nil)

	assert.Contains(t, string(body), expectedText, failText)
}

func postContainsText(t *testing.T, queryUrl, postBody, expectedText, failText string) string {
	parsedUrl, _ := url.Parse(queryUrl)

	resp, err := http.Post(parsedUrl.String(), "application/json", strings.NewReader(postBody))
	if err != nil {
		fmt.Println(err.Error())
	}
	assert.Equal(t, err, nil)

	body, err := io.ReadAll(resp.Body)
	assert.Equal(t, err, nil)

	bodyString := string(body)
	assert.Contains(t, bodyString, expectedText, failText)
	return bodyString
}
