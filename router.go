package x

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strings"
)

type router struct {
	method    string
	path      string
	functions []contextFunction
}

func (r *router) execute(w http.ResponseWriter, req *http.Request) {
	defer func() {
		err := recover()
		switch v := err.(type) {
		case xError:
			return
		case writeError:
			return
		default:
			if v != nil {
				panic(v)
			}
			return
		}
	}()

	// 检查session状态
	var sess *session
	if openSession {
		sess = checkSession(w, req)
	}

	// 创建context
	var ctx = Context{
		request:        req,
		responseWriter: w,
		query:          req.URL.Query(),
		session:        sess,
	}

	// 解析请求体
	parseBody(&ctx, req)

	if len(r.functions) > 1 {
		ctx.tempFunctions = r.functions[1:]
	}

	r.functions[0](ctx)
	return
}

func Use(f ...contextFunction) {
	globalBefore = append(globalBefore, f...)
}

func After(f ...contextFunction) {
	globalAfter = append(globalAfter, f...)
}

func parseBody(ctx *Context, req *http.Request) {
	m := toUpper(req.Method)
	contentType := req.Header.Get("Content-Type")
	if m == _POST || m == _PUT || m == _PATCH {
		if contentType == "application/json" {
			decoder := json.NewDecoder(req.Body)
			params := make(map[string]interface{})
			decoder.Decode(&params)
			ctx.json = params
		} else if contentType == "application/x-www-form-urlencoded" {
			err := req.ParseForm()
			if err != nil {
				ctx.files = make(map[string][]*multipart.FileHeader)
			} else {
				ctx.body = req.Form
				mf := req.MultipartForm
				if mf == nil {
					ctx.files = make(map[string][]*multipart.FileHeader)
				} else {
					ctx.files = mf.File
				}
			}
		} else if strings.HasPrefix(contentType, "multipart/form-data") {
			err := req.ParseMultipartForm(10 << 22)
			if err != nil {
				ctx.files = make(map[string][]*multipart.FileHeader)
			} else {
				mf := req.MultipartForm
				ctx.body = mf.Value
				if mf == nil {
					ctx.files = make(map[string][]*multipart.FileHeader)
				} else {
					ctx.files = mf.File
				}
			}
		}
	}
}


func StaticServer(prefix string, dir string) {
	staticServer[prefix] = dir
}

func StrictSlash(flag bool) {
	if flag {
		banSlash = false
	}
	strictSlash = flag
}

func BanSlash(flag bool) {
	if flag {
		strictSlash = false
	}
	banSlash = flag
}
