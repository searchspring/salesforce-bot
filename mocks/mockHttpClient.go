package mocks

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

// Helpful Blog https://www.thegreatcodeadventure.com/mocking-http-requests-in-golang/J

type HttpClient struct{}

func (httpClient HttpClient) Do(req *http.Request) (*http.Response, error) {
	// default to a 404. if we find our matching URLs for our tests, we'll return some good stuff
	response := http.Response{
		Status:     "404",
		StatusCode: 404,
		Body:       ioutil.NopCloser(bytes.NewReader(make([]byte, 0, 0))),
	}

	if strings.HasSuffix(req.URL.Path, "/q8q4eu/exclusionStats") {
		stats := map[string]interface{}{
			"hello":            "goodbye",
			"tags=Overexposed": 18,
		}
		asBytes, _ := json.Marshal(stats)

		response.Body = ioutil.NopCloser(bytes.NewReader(asBytes))
		response.StatusCode = 200
		response.Status = "200"
	} else if strings.HasSuffix(req.URL.Path, "/q8q4eu/status") {
		status := map[string]interface{}{
			"overallStatus":                 "Completed. Version A",
			"lastExtractionDurationMinutes": 2,
		}
		asBytes, _ := json.Marshal(status)

		response.Body = ioutil.NopCloser(bytes.NewReader(asBytes))
		response.StatusCode = 200
		response.Status = "200"
	}
	return &response, nil
}
