package api

import (
	"testing"
	"time"

	"github.com/searchspring/nebo/common"
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

func TestTimestamp(t *testing.T) {
	require.Equal(t, "2020-10-29-14-08", timestamp(time.Unix(1603980505, 0)))
}
