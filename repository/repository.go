// Package repository contains repository methods abstractions.
package repository

import (
	"errors"
	"strings"
)

// Errors.
var (
	ErrUniqueViolation  = errors.New("resource already existing")
	ErrResourceNotFound = errors.New("resource not found")
)

// StripWhitespaces strips out all duplicated whitespaces in a string.
func StripWhitespaces(s string) string {
	if s = regex.ReplaceAllString(strings.TrimSpace(s), " "); s != " " {
		return s
	}

	return ""
}
