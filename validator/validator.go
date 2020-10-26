package validator

// ContainsEmptyString returns true if any of the string variables provided are blank
func ContainsEmptyString(vars ...string) bool {
	for _, v := range vars {
		if v == "" {
			return true
		}
	}
	return false
}
