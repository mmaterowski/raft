package guard

func AgainstEmptyString(s string) bool {
	return len(s) > 0
}

func AgainstNegativeValue(i int) bool {
	return i < 0
}
