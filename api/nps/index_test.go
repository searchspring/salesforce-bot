package nps

import (
	"log"
	"net/http/httptest"
	"testing"
	"os"

	"github.com/stretchr/testify/require"
	"github.com/searchspring/nebo/salesforce"
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
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt", nil), &SlackDAOFake{}, &salesforce.DAOfake{})
	require.Equal(t, 500, w.Result().StatusCode)
}

func TestHandlerSendSlackMessage(t *testing.T) {
	os.Setenv("DEV_MODE", "development")
	defer os.Setenv("DEV_MODE", "")
	w := httptest.NewRecorder()
	slack := &SlackDAOFake{}
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt&rating=10&email=matt@smith.test&website=mattsmith.test", nil), slack, &salesforce.DAOfake{})
	require.Equal(t, []string{"", ""}, slack.getValues())
}

func TestSalesforceQuery(t *testing.T) {
	sfdao := &salesforce.DAOfake{}
	query := "test"
	response, err := sfdao.NPSQuery(query)
	if err != nil {
		log.Println(err)
		t.Fail()
	}
	require.Equal(t, query, response[0].Manager)
}

func TestSalesforceSearchKey(t *testing.T) {
	os.Setenv("DEV_MODE", "development")
	defer os.Setenv("DEV_MODE", "")
	w := httptest.NewRecorder()
	sfdao := &salesforce.DAOfake{}
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt&rating=10&email=matt@smith.test&website=mattsmith.test%20(2003)", nil), &SlackDAOFake{}, sfdao)
	require.Equal(t, "mattsmith.test", sfdao.GetSearchKey())
}

