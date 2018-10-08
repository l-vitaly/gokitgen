package utils

import (
	"strings"
)

func LcFirst(v string) string {
	return strings.ToLower(v[:1]) + v[1:]
}

func UcFirst(v string) string {
	return strings.ToUpper(v[:1]) + v[1:]
}
