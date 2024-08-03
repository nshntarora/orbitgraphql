package utils

func StringArrayContainsString(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func ArrayContains(arr []any, el any) bool {
	for _, a := range arr {
		if a == el {
			return true
		}
	}
	return false
}
