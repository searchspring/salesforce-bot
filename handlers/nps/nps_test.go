package nps

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/searchspring/nebo/common"
	"github.com/stretchr/testify/require"
	"github.com/searchspring/nebo/mocks"
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
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt", nil), &mocks.SlackDAO{}, &mocks.MetabaseDAO{})
	require.Equal(t, 500, w.Result().StatusCode)
}

func TestHandlerSendSlackMessage(t *testing.T) {
	os.Setenv("DEV_MODE", "development")
	defer os.Setenv("DEV_MODE", "")
	w := httptest.NewRecorder()
	slack := &mocks.SlackDAO{}
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt&rating=10&email=matt@smith.test&website=mattsmith.test", nil), slack, &mocks.MetabaseDAO{})
	require.Equal(t, []string{"", ""}, slack.GetValues())
}

func TestMetabaseQuery(t *testing.T) {
	mbdao := &mocks.MetabaseDAO{}
	query := "tester"
	response, err := mbdao.QueryNPS(query)
	if err != nil {
		log.Println(err)
		t.Fail()
	}
	require.Equal(t, query, response.Manager)
}

func TestMetabaseSearchKey(t *testing.T) {
	os.Setenv("DEV_MODE", "development")
	defer os.Setenv("DEV_MODE", "")
	w := httptest.NewRecorder()
	mbdao := &mocks.MetabaseDAO{}
	SendNPSMessage(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt&rating=10&email=matt@smith.test&website=mattsmith.test%20(2003)", nil), &mocks.SlackDAO{}, mbdao)
	require.Equal(t, "mattsmith", mbdao.GetSearchKey())
}

