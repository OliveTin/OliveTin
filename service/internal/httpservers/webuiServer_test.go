package httpservers

import (
	"fmt"
	config "github.com/OliveTin/OliveTin/internal/config"
	"github.com/stretchr/testify/assert"	
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGetWebuiDir(t *testing.T) {
	originalDir, err := os.Getwd()
	assert.Equal(t, nil, err)
	os.Chdir("../../") // go test sets the cwd to "httpservers" by default
	defer os.Chdir(originalDir)

	cfg = config.DefaultConfig()

	dir := findWebuiDir()

	assert.Equal(t, "../webui/", dir, "Finding the webui dir")
}

func TestGetBaseUrl(t *testing.T) {
	originalDir, err := os.Getwd()
	assert.Equal(t, nil, err)
	os.Chdir("../../") // go test sets the cwd to "httpservers" by default
	defer os.Chdir(originalDir)
	cfg = config.DefaultConfig()

	type testCase struct {
		TestName            string
		ExternalRestAddress string
		Subpath             string
		expectedBaseUrl     string
		expectedCookiePath  string
	}

	testCases := []testCase{
		{
			TestName:            "Test default configuration",
			ExternalRestAddress: ".",
			Subpath:             "",
			expectedBaseUrl:     "/",
			expectedCookiePath:  "/",
		},
		{
			TestName:            "Test where an external rest address is used",
			ExternalRestAddress: "localhost:1337",
			Subpath:             "/subpath",
			expectedBaseUrl:     "localhost:1337/subpath/",
			expectedCookiePath:  "/subpath",
		},
		{
			TestName:            "Test where an external rest address is used with trailing suffix",
			ExternalRestAddress: "localhost:1337",
			Subpath:             "/subpath/",
			expectedBaseUrl:     "localhost:1337/subpath/",
			expectedCookiePath:  "/subpath",
		},
		{
			TestName:            "Test where an external rest address is used with trailing suffixes everywhere",
			ExternalRestAddress: "localhost:1337/",
			Subpath:             "/subpath/",
			expectedBaseUrl:     "localhost:1337/subpath/",
			expectedCookiePath:  "/subpath",
		},
		{
			TestName:            "Test where an external rest address is used with an egregious amount of trailing suffixes",
			ExternalRestAddress: "localhost:1337/",
			Subpath:             "///subpath/////",
			expectedBaseUrl:     "localhost:1337/subpath/",
			expectedCookiePath:  "/subpath",
		},
		{
			TestName:            "Test with a different port",
			ExternalRestAddress: "localhost:1300",
			Subpath:             "/subpath/",
			expectedBaseUrl:     "localhost:1300/subpath/",
			expectedCookiePath:  "/subpath",
		},
		{
			TestName:            "Test with a different port and no subpath",
			ExternalRestAddress: "localhost:1300",
			Subpath:             "",
			expectedBaseUrl:     "localhost:1300",
			expectedCookiePath:  "/",
		},
		{
			TestName:            "Test the default external rest address and a subpath",
			ExternalRestAddress: ".",
			Subpath:             "/subpath",
			expectedBaseUrl:     "/subpath/",
			expectedCookiePath:  "/subpath",
		},
	}

	// Check the default configuration
	url := baseURL()
	assert.Equal(t, "/", url)

	for _, testCase := range testCases {
		t.Run(testCase.TestName, func(t *testing.T) {
			// Create a temporary copy of the configuration for this test
			testCfg := *cfg
			testCfg.ExternalRestAddress = testCase.ExternalRestAddress
			testCfg.Subpath = testCase.Subpath

			// Store original config and restore after test
			origCfg := cfg
			cfg = &testCfg
			defer func() { cfg = origCfg }()

			url := baseURL()
			assert.Equal(t, testCase.expectedBaseUrl, url, "Test \"%s\" failed", testCase.TestName)

			rr := httptest.NewRecorder()

			// Confirm that the indexHtml can be re-written to configure the base href
			serveIndexHtmlWithBasePath(rr)

			if status := rr.Code; status != http.StatusOK {
				t.Errorf("expected status %v, got %v", http.StatusOK, status)
			}
			contentType := rr.Header().Get("content-type")
			assert.Equal(t, "text/html", contentType)
			assert.Contains(t, rr.Body.String(), "<title>OliveTin</title>")
			expectedBaseHref := fmt.Sprintf("<base href=\"%s\">", testCase.expectedBaseUrl)
			assert.Contains(t, rr.Body.String(), expectedBaseHref)

			path := getCookiePath()

			assert.Equal(t, testCase.expectedCookiePath, path, "Test \"%s\" failed", testCase.TestName)
		})
	}
}
