package x

import (
	"regexp"
	"time"
)

type (
	Map map[string]interface{}
)

type (
	contextFunction func(Context)
)

var (
	globalCors   = false
	openSession  = false
	singleSlash = true
	strictSlash = false
	banSlash = false
	sessOption   = sessionOption{"sid", 0, false}
	sessionStore = map[string]*session{}
	routerMap    = map[string]map[string]*router{
		_GET:    map[string]*router{},
		_PUT:    map[string]*router{},
		_POST:   map[string]*router{},
		_PATCH:  map[string]*router{},
		_DELETE: map[string]*router{},
	}
	methods      = []string{_GET, _PUT, _POST, _PATCH, _DELETE, _OPTIONS}
	globalBefore = []contextFunction{}
	globalAfter  = []contextFunction{}
	sec, _       = time.ParseDuration("1s")
	staticServer = map[string]string{}
	_smartSlash, _ = regexp.Compile("/+")
	lastSlash, _ = regexp.Compile("/+$")
)

const (
	_GET     = "GET"
	_PUT     = "PUT"
	_POST    = "POST"
	_PATCH   = "PATCH"
	_DELETE  = "DELETE"
	_OPTIONS = "OPTIONS"
)

const (
	HTTP_OK                    = 200
	HTTP_BAD_REQUEST           = 400
	HTTP_UNAUTHORIZED          = 401
	HTTP_NOT_FOUND             = 404
	HTTP_INTERNAL_SERVER_ERROR = 500
)
