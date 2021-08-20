package boost

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHelpTextIncludesAllCommands(t *testing.T) {
	boostHelpText := HelpText()
	for _, command := range SlackCommands {
		require.Contains(t, boostHelpText, command)
	}
}

