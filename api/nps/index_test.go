package nps

import (
	"fmt"
	"log"
	"net/http/httptest"
	"testing"
	"os"

	"github.com/stretchr/testify/require"
)

func TestFindBlankEnvVars(t *testing.T) {
	testVars := envVars{
		DevMode: "test",
	}
	blanks := findBlankEnvVars(testVars)
	for _, b := range blanks {
		require.NotEqual(t, "DevMode", b)
	}
}

func TestHandlerMissingEnvVars(t *testing.T) {
	w := httptest.NewRecorder()
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt", nil), &SlackDAOFake{})
	require.Equal(t, 500, w.Result().StatusCode)
}

func TestHandlerSendSlackMessage(t *testing.T) {
	os.Setenv("DEV_MODE", "development")
	w := httptest.NewRecorder()
	slack := &SlackDAOFake{}
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt&rating=10&email=matt@smith.test&website=mattsmith.test", nil), slack)
	require.Equal(t, []string{}, slack.getValues())
	defer os.Setenv("DEV_MODE", "")
}

func TestParseUrl(t *testing.T) {
	urlString, err := parseUrl(httptest.NewRequest("GET", "localhost:3000/nps?name=Matt&email=matt@smith.test&website=mattsmith.test&feedback=Perfect", nil))
	fmt.Println(urlString)
	if err != nil {
		log.Println(err)
		t.Fail()
	}
}
