package utils

import "encoding/json"

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

func ConvertMapToString(m map[string]interface{}) string {
	br, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(br)
}
