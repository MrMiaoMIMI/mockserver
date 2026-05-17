package dao

import "strings"

func isDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "Error 1062") || strings.Contains(strings.ToLower(message), "duplicate entry")
}
