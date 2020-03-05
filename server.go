package x

import "net/http"

func Run(addr string) error {
	return http.ListenAndServe(addr, &mux{})
}