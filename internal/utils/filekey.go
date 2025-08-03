package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func MakeFileKey(pkg, version, fileName string) string {
	safePkg := url.PathEscape(strings.ToLower(pkg))
	safeVer := url.PathEscape(strings.ToLower(version))
	safeFile := url.PathEscape(fileName)
	return fmt.Sprintf("uploads/%s/%s/%s", safePkg, safeVer, safeFile)
}
