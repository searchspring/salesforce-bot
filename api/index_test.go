package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindBlankEnvVars(t *testing.T) {
	testVars := envVars{
		DevMode:                     "test",
		NxUser:                      "",
		SlackVerificationToken:      "test",
		SlackOauthToken:             "test",
		SfURL:                       "test",
		SfUser:                      "test",
		SfPassword:                  "test",
		SfToken:                     "test",
		NxPassword:                  "test",
		GcpServiceAccountEmail:      "test",
		GcpServiceAccountPrivateKey: "test",
		GdriveFireDocFolderID:       "test",
	}
	blanks := findBlankEnvVars(testVars)
	require.Equal(t, "NxUser", blanks[0])
	require.Equal(t, 1, len(blanks))
	fmt.Println(blanks)
}
