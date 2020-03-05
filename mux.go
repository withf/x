package x

import (
	"net/http"
	"os"
	"strings"
)

type mux struct{}

func (m *mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	// 全局跨域访问
	if globalCors {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
		if method == _OPTIONS {
			return
		}
	}

	// 静态资源检查
	if method == "GET" {
		for k, v := range staticServer {
			if strings.HasPrefix(path, k) {
				name := v + "/" + path[len(k):]
				f, err := os.Open(name)
				if err != nil {
					w.WriteHeader(HTTP_NOT_FOUND)
					return
				}
				fi, err := f.Stat()
				if err != nil {
					w.WriteHeader(HTTP_NOT_FOUND)
					return
				}
				b := make([]byte, fi.Size())
				_, err = f.Read(b)
				if err != nil {
					w.WriteHeader(HTTP_INTERNAL_SERVER_ERROR)
				}
				err = f.Close()
				if err != nil {
					w.WriteHeader(HTTP_INTERNAL_SERVER_ERROR)
				}
				_, err = w.Write(b)
				if err != nil {
					w.WriteHeader(HTTP_INTERNAL_SERVER_ERROR)
				}
				return
			}
		}
	}

	// 执行对应路由的函数
	r, ok := routerMap[method][path]
	if ok {
		r.execute(w, req)
	} else {
		w.WriteHeader(HTTP_NOT_FOUND)
	}
}
