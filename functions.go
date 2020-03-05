package x

import (
	"errors"
	"strings"
)

func toUpper(s string) string {
	return strings.ToUpper(s)
}

func checkMethod(m string) error {
	m = toUpper(m)
	for _, v := range methods {
		if m == v {
			return nil
		}
	}
	return errors.New("the request method is incorrect")
}

func _singleSlash(s string) string {
	if singleSlash {
		s = _smartSlash.ReplaceAllString(s, "/")
	}
	s = _strictSlash(s)
	s = _banSlash(s)
	return s
}

func _strictSlash(s string) string {
	if strictSlash {
		s += "/"
	}
	return s
}

func _banSlash(s string) string {
	if banSlash {
		s = lastSlash.ReplaceAllString(s, "")
	}
	return s
}

func combineSlash(s string, p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return _singleSlash(s + p)
}
