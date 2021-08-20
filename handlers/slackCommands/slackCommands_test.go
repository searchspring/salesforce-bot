package api

import (
	"github.com/stretchr/testify/assert"
	"strings"
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

func TestFormatMapResponse(t *testing.T) {
	exclusionStats := map[string]interface{}{
		"thing1": "hummus",
		"thing2": 23,
		"thing3": false,
	}
	formattedForSlack := FormatMapResponse(exclusionStats)
	assert.Truef(t, strings.HasPrefix(formattedForSlack, "```"), "Not the right prefix")
	assert.Truef(t, strings.HasSuffix(formattedForSlack, "\n```"), "Not the right suffix")

	expectedString := "```thing1: hummus\nthing2: 23\nthing3: false\n```"
	assert.Equal(t, formattedForSlack, expectedString)
}

func TestGetMeetLink(t *testing.T) {
	assert.Equal(t, "g.co/meet/pickles-and-hummus", GetMeetLink("pickles and hummus"))

	randomlyGeneratedName := GetMeetLink("")
	const prefix = "g.co/meet/"
	assert.Truef(t, strings.HasPrefix(randomlyGeneratedName, prefix), "Starts with prefix")

	// meetName should be three words separated by hyphens like "generously-playful-peanut"
	meetName := strings.Split(randomlyGeneratedName, prefix)[1]
	assert.Equal(t, len(strings.Split(meetName, "-")), 3)
}
