package listSites

import (
	"testing"

	common "github.com/searchspring/nebo/api/config"
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
