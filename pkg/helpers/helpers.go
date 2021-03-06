package helpers

import (
	"regexp"
	"unicode"
)

var emailRegex = regexp.MustCompile("^\\w+([\\.-]?\\w+)*@\\w+([\\.-]?\\w+)*(\\.\\w{2,3})+$")

func IsAlphaNumeric(str string) bool {
	for _, char := range str {
		if !(unicode.IsLetter(char) || unicode.IsNumber(char)) {
			return false
		}
	}
	return true
}

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func ContainsStr(haystack []string, needle string, transforms ...func(string) string) bool {
	for _, item := range haystack {
		for _, t := range transforms {
			item = t(item)
		}
		if item == needle {
			return true
		}
	}
	return false
}
