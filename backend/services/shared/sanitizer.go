package shared

import "strings"

func EscapeFilename(filename string) string {
	f := strings.ReplaceAll(filename, "\\", ":")
	f = strings.ReplaceAll(f, "/", ":")
	return f
}
