package listSites

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/searchspring/nebo/common"
	mocks "github.com/searchspring/nebo/mocks"
	"github.com/stretchr/testify/require"
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

func TestGetSitesList(t *testing.T) {
	os.Setenv("DEV_MODE", "development")
	defer os.Setenv("DEV_MODE", "")
	w := httptest.NewRecorder()
	metabaseDAO := &mocks.MetabaseDAO{}
	GetSitesList(w, httptest.NewRequest("GET", "localhost:3000/listSites", nil), metabaseDAO)
	require.Equal(t, 201, w.Result().StatusCode)
}
