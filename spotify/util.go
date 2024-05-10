package spotify

import (
	"fmt"
	"strings"
)

func IDFromURI(uri string) (string, error) {
	segments := strings.Split(uri, ":")
	if len(segments) != 3 {
		return "", fmt.Errorf("bad uri format (%s)", uri)
	}
	return segments[2], nil
}

func IDFromURIMust(uri string) string {
	segments := strings.Split(uri, ":")
	if len(segments) != 3 {
		return ""
	}
	return segments[2]
}
