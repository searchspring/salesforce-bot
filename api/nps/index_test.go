package nps

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/searchspring/nebo/api/config"
	"github.com/stretchr/testify/require"
	"github.com/searchspring/nebo/api/tests"
)

func TestFindBlankEnvVars(t *testing.T) {
	testVars := common.EnvVars{
		DevMode: "test",
	}
	blanks := common.FindBlankEnvVars(testVars)
	for _, b := range blanks {
		require.NotEqual(t, "DevMode", b)
	}
}

func TestHandlerMissingEnvVars(t *testing.T) {
	w := httptest.NewRecorder()
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt", nil), &mocks.SlackDAO{}, &mocks.SalesforceDAO{})
	require.Equal(t, 500, w.Result().StatusCode)
}

func TestHandlerSendSlackMessage(t *testing.T) {
	os.Setenv("DEV_MODE", "development")
	defer os.Setenv("DEV_MODE", "")
	w := httptest.NewRecorder()
	slack := &mocks.SlackDAO{}
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt&rating=10&email=matt@smith.test&website=mattsmith.test", nil), slack, &mocks.SalesforceDAO{})
	require.Equal(t, []string{"", ""}, slack.GetValues())
}

func TestSalesforceQuery(t *testing.T) {
	sfdao := &mocks.SalesforceDAO{}
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
	sfdao := &mocks.SalesforceDAO{}
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt&rating=10&email=matt@smith.test&website=mattsmith.test%20(2003)", nil), &mocks.SlackDAO{}, sfdao)
	require.Equal(t, "mattsmith", sfdao.GetSearchKey())
}

