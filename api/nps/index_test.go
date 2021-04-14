package nps

import (
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	//"time"
	//"github.com/stretchr/testify/require"
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

// test that response contains the correct fields
func TestHandlerMissingEnvVars(t *testing.T) {
	w := httptest.NewRecorder()
	Handler(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt", nil))
	require.Equal(t, 500, w.Result().StatusCode)

}

func TestHandlerSendSlackMessage(t *testing.T) {
	os.Setenv("DEV_MODE", "development")
	defer os.Setenv("DEV_MODE", "")
	w := httptest.NewRecorder()
	Handler(w, httptest.NewRequest("GET", "localhost:3000/nps?name=Matt&rating=10&email=matt@smith.test&website=mattsmith.test&test=true", nil))
	require.Equal(t, []string{"", "C01TWG8D6CC"}, slackDAO.getValues())
}

func TestParseUrl(t *testing.T) {
	urlString, err := parseUrl(httptest.NewRequest("GET", "localhost:3000/nps?name=Matt&email=matt@smith.test&website=mattsmith.test&feedback=Perfect&test=true", nil))
	fmt.Println(urlString)
	if err != nil {
		log.Println(err)
		t.Fail()
	}

}
