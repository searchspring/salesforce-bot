package api

import (
	"fmt"
	"testing"
)

func TestFindBlankEnvVars(t *testing.T) {
	testVars := envVars{
		DevMode: "test",
		NxUser:  "",
	}
	blanks := findBlankEnvVars(testVars)
	fmt.Println(blanks)
}
