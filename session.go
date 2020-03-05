package x

import (
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strings"
	"time"
)

type session struct {
	id         string
	values     map[string]interface{}
	expiration time.Time
}

type sessionOption struct {
	key        string
	Expiration time.Duration
	AutoFlash  bool
}

func OpenSession(key string, expiration time.Duration, autoFlash bool) {
	openSession = true
	sessOption.key = key
	sessOption.Expiration = expiration
	sessOption.AutoFlash = autoFlash
}

func (s *session) Set(name string, v interface{}) {
	s.values[name] = v
}

func (s *session) Get(name string) interface{} {
	return s.values[name]
}

func (s *session) Destroy() {
	delete(sessionStore, s.id)
}

func (s *session) Clear() {
	s.values = map[string]interface{}{}
}


func checkSession(w http.ResponseWriter, req *http.Request) *session {
	sess, exist := getSession(w, req)
	if !exist {
		sessionStore[sess.id] = sess
	} else if sess.expiration.Before(time.Now()) {
		sess.Clear()
		sess.expiration = time.Now().Add(sessOption.Expiration * sec)
	} else if sessOption.AutoFlash {
		sess.expiration = time.Now().Add(sessOption.Expiration * sec)
	}
	return sess
}

func getSession(w http.ResponseWriter, req *http.Request) (*session, bool) {
	var sessionID string
	cookiePtr, err := req.Cookie(sessOption.key)
	if err == nil && cookiePtr != nil {
		sessionID = cookiePtr.Value
	}
	sess := sessionStore[sessionID]
	exist := true
	if sess == nil {
		exist = false
		cookie := newCookie(req)
		http.SetCookie(w, cookie)
		t := time.Now().Add(sessOption.Expiration * sec)
		sess = &session{values: map[string]interface{}{}, id: cookie.Value, expiration: t}
	}
	return sess, exist
}

func newCookie(req *http.Request) *http.Cookie {
	return &http.Cookie{
		Name:   sessOption.key,
		Value:  uuid.NewV4().String(),
		Path:   "/",
		Domain: getDomain(req),
		MaxAge: 0,
	}
}

func getDomain(req *http.Request) string {
	requestDomain := req.Host
	if portIdx := strings.IndexByte(requestDomain, ':'); portIdx > 0 {
		requestDomain = requestDomain[0:portIdx]
	}
	return requestDomain
}
