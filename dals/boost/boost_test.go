package boost

import (
	"github.com/searchspring/nebo/common"
	"github.com/searchspring/nebo/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHelpTextIncludesAllCommands(t *testing.T) {
	boostHelpText := HelpText()
	for _, command := range SlackCommands {
		require.Contains(t, boostHelpText, command)
	}
}

func TestGetExclusionStats(t *testing.T) {
	exclusions := HandleGetExclusionStatsRequest("q8q4eu", common.NewClient(mocks.HttpClient{}))
	assert.Equal(t, "goodbye", exclusions["hello"])
	assert.Equal(t, float64(18), exclusions["tags=Overexposed"])

	exclusions = HandleGetExclusionStatsRequest("fakeNewsSiteId", common.NewClient(mocks.HttpClient{}))
	assert.Emptyf(t, exclusions, "Go, why require this message")
}

func TestGetStatus(t *testing.T) {
	status := HandleGetStatusRequest("q8q4eu", common.NewClient(mocks.HttpClient{}))
	assert.Equal(t, "Completed. Version A", status["overallStatus"])
	assert.Equal(t, float64(2), status["lastExtractionDurationMinutes"])

	status = HandleGetStatusRequest("notGonnaFindThisOne", common.NewClient(mocks.HttpClient{}))
	assert.Emptyf(t, status, "Why must you require this")
}
