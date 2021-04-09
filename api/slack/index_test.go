package api

import (
	"testing"
	"time"

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

func TestTimestamp(t *testing.T) {
	require.Equal(t, "2020-10-29-14-08", timestamp(time.Unix(1603980505, 0)))
}
