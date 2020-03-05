package x

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Context struct {
	request        *http.Request
	responseWriter http.ResponseWriter
	query          url.Values
	json           map[string]interface{}
	body           url.Values
	files          map[string][]*multipart.FileHeader
	tempFunctions  []contextFunction
	*session
}

func checkCode(code ...int) int {
	var status int
	if len(code) == 0 {
		status = HTTP_OK
	} else {
		status = code[0]
	}
	return status
}

func (c *Context) Next() {
	if len(c.tempFunctions) > 0 {
		t := c.tempFunctions[0]
		if len(c.tempFunctions) > 1 {
			c.tempFunctions = c.tempFunctions[1:]
		}
		t(*c)
	}
}

func (c *Context) SetHeader(name, value string) {
	c.responseWriter.Header().Set(name, value)
}

func (c *Context) SetHeaders(m Map) {
	for k, v := range m {
		c.SetHeader(k, v.(string))
	}
}

func (c *Context) StatusCode(code int) {
	c.responseWriter.WriteHeader(code)
}

func (c *Context) Redirect(url string, statusCode int) {
	http.Redirect(c.responseWriter, c.request, url, statusCode)
}

func (c *Context) Request() *http.Request {
	return c.request
}

func (c *Context) ResponseWriter() http.ResponseWriter {
	return c.responseWriter
}

func (c *Context) RawQuery() url.Values {
	return c.query
}

func (c *Context) RawBody() url.Values {
	return c.body
}

func (c *Context) RawJson() map[string]interface{} {
	return c.json
}

func (c *Context) RawFiles() map[string][]*multipart.FileHeader {
	return c.files
}

func (c *Context) SendJSON(v interface{}, code ...int) {
	status := checkCode(code...)
	c.SetHeader("Content-Type", "application/json")
	c.StatusCode(status)
	b, err := json.Marshal(v)
	if err != nil {
		panic(writeError{})
	}
	_, err = c.responseWriter.Write(b)
	if err != nil {
		panic(writeError{})
	}
}

func (c *Context) SendJSONReturn(v interface{}, code ...int) {
	status := checkCode(code...)
	c.SendJSON(v, status)
	panic(xError{})
}

func (c *Context) Write(b []byte, code ...int) {
	status := checkCode(code...)
	c.StatusCode(status)
	_, err := c.responseWriter.Write(b)
	if err != nil {
		panic(writeError{})
	}
}

func (c *Context) WriteReturn(b []byte, code ...int) {
	status := checkCode(code...)
	c.Write(b, status)
	panic(xError{})
}

func (c *Context) WriteString(str string, code ...int) {
	status := checkCode(code...)
	c.StatusCode(status)
	_, err := c.responseWriter.Write([]byte(str))
	if err != nil {
		panic(writeError{})
	}
}

func (c *Context) WriteStringReturn(str string, code ...int) {
	status := checkCode(code...)
	c.WriteString(str, status)
	panic(xError{})
}

func (c *Context) Check(errOrBool interface{}, statusCode int, vs ...interface{}) {
	if errOrBool == nil {
		return
	}
	var v interface{}
	if len(vs) != 0 {
		v = vs[0]
	}
	switch e := errOrBool.(type) {
	case error:
		if e != nil {
			if v != nil {
				switch v.(type) {
				case string:
					c.WriteStringReturn(v.(string), statusCode)
				default:
					c.SendJSONReturn(v, statusCode)
				}
			} else {
				c.StatusCode(statusCode)
				panic(xError{})
			}
		}
	case bool:
		if !e {
			if v != nil {
				switch v.(type) {
				case string:
					c.WriteStringReturn(v.(string), statusCode)
				default:
					c.SendJSONReturn(v, statusCode)
				}
			} else {
				c.StatusCode(statusCode)
				panic(xError{})
			}
		}
	default:
		panic("The first param of function Check must be error or bool type.")
	}
}

func (c *Context) QueryValues(name string) ([]string, bool) {
	q, ok := c.query[name]
	return q, ok
}

func (c *Context) queryInt(name string) (int, bool, error) {
	q, ok := c.query[name]
	if ok {
		r, err := strconv.Atoi(q[0])
		return r, ok && err == nil, err
	}
	return 0, ok, nil
}

func (c *Context) QueryInt(name string) (int, bool) {
	v, exist, err := c.queryInt(name)
	return v, exist && err == nil
}

func (c *Context) QueryIntSlice(name string) ([]int, bool) {
	var r []int
	q, ok := c.query[name]
	if ok {
		r = make([]int, len(q))
		for index, v := range q {
			a, err := strconv.Atoi(v)
			if err != nil {
				return r, false
			}
			r[index] = a
		}
	}
	return r, ok
}

func (c *Context) QueryIntDefault(name string, n int) int {
	r := n
	q, ok := c.query[name]
	if ok {
		r, _ = strconv.Atoi(q[0])
	}
	return r
}

func (c *Context) queryFloat64(name string) (float64, bool, error) {
	q, ok := c.query[name]
	if ok {
		r, err := strconv.ParseFloat(q[0], 64)
		return r, ok && err == nil, err
	}
	return 0, ok, nil
}

func (c *Context) QueryFloat64(name string) (float64, bool) {
	v, exist, err := c.queryFloat64(name)
	return v, exist && err == nil
}

func (c *Context) QueryFloat64Slice(name string) ([]float64, bool) {
	var r []float64
	q, ok := c.query[name]
	if ok {
		r = make([]float64, len(q))
		for index, v := range q {
			a, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return r, false
			}
			r[index] = a
		}
	}
	return r, ok
}

func (c *Context) QueryFloat64Default(name string, n float64) float64 {
	r := n
	q, ok := c.query[name]
	if ok {
		r, _ = strconv.ParseFloat(q[0], 64)
	}
	return r
}

func (c *Context) QueryString(name string) (string, bool) {
	q, ok := c.query[name]
	if ok {
		return q[0], ok
	}
	return "", ok
}

func (c *Context) QueryStringSlice(name string) ([]string, bool) {
	q, ok := c.query[name]
	return q, ok
}

func (c *Context) QueryStringDefault(name string, s string) (string, bool) {
	r := s
	q, ok := c.query[name]
	if ok {
		r = q[0]
	}
	return r, ok
}

func (c *Context) queryBool(name string) (bool, bool, error) {
	q, ok := c.query[name]
	if ok {
		r, err := strconv.ParseBool(q[0])
		return r, ok && err == nil, err
	}
	return false, ok, nil
}

func (c *Context) QueryBool(name string) (bool, bool) {
	v, exist, err := c.queryBool(name)
	return v, exist && err == nil
}

func (c *Context) QueryBoolDefault(name string, b bool) bool {
	r := b
	q, ok := c.query[name]
	if ok {
		a, err := strconv.ParseBool(q[0])
		r = a && err == nil
	}
	return r
}

func (c *Context) PostValue(name string) (string, bool) {
	r, ok := c.body[name]
	if !ok {
		return "", ok
	}
	return r[0], ok
}

func (c *Context) PostValueTrim(name string) (string, bool) {
	r, ok := c.body[name]
	if !ok {
		return "", ok
	}
	return strings.Trim(r[0], " "), ok
}

func (c *Context) PostValueDefault(name string, v string) string {
	r, ok := c.body[name]
	if !ok {
		return v
	}
	return r[0]
}

func (c *Context) PostValueSlice(name string) ([]string, bool) {
	t, ok := c.body[name]
	return t, ok
}

func (c *Context) postValueInt(name string) (int, bool, error) {
	q, ok := c.body[name]
	if ok {
		r, err := strconv.Atoi(q[0])
		return r, ok && err == nil, err
	}
	return 0, ok, nil
}

func (c *Context) PostValueInt(name string) (int, bool) {
	v, exist, err := c.postValueInt(name)
	return v, exist && err == nil
}

func (c *Context) PostValueIntDefault(name string, v int) int {
	q, ok := c.body[name]
	if ok {
		r, err := strconv.Atoi(q[0])
		if err != nil {
			return v
		}
		return r
	}
	return v
}

func (c *Context) PostValueIntSlice(name string) ([]int, bool) {
	var r []int
	q, ok := c.body[name]
	if !ok {
		r := make([]int, len(q))
		for index, v := range q {
			t, err := strconv.Atoi(v)
			if err != nil {
				return r, false
			}
			r[index] = t
		}
		return r, true
	}
	return r, false
}

func (c *Context) postValueFloat64(name string) (float64, bool, error) {
	q, ok := c.body[name]
	if ok {
		r, err := strconv.ParseFloat(q[0], 64)
		return r, ok && err == nil, err
	}
	return 0, ok, nil
}

func (c *Context) PostValueFloat64(name string) (float64, bool) {
	v, exist, err := c.postValueFloat64(name)
	return v, exist && err == nil
}

func (c *Context) PostValueFloat64Default(name string, v float64) float64 {
	q, ok := c.body[name]
	if ok {
		r, err := strconv.ParseFloat(q[0], 64)
		if err != nil {
			return v
		}
		return r
	}
	return v
}

func (c *Context) PostValueFloat64Slice(name string) ([]float64, bool) {
	var r []float64
	q, ok := c.body[name]
	if !ok {
		r := make([]float64, len(q))
		for index, v := range q {
			t, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return r, false
			}
			r[index] = t
		}
		return r, true
	}
	return r, false
}

func (c *Context) postValueBool(name string) (bool, bool, error) {
	q, ok := c.body[name]
	if ok {
		r, err := strconv.ParseBool(q[0])
		return r, ok && err == nil, err
	}
	return false, ok, nil
}

func (c *Context) PostValueBool(name string) (bool, bool) {
	v, exist, err := c.postValueBool(name)
	return v, exist && err == nil
}

func (c *Context) PostValueBoolDefault(name string, v bool) bool {
	q, ok := c.body[name]
	if ok {
		r, err := strconv.ParseBool(q[0])
		if err != nil {
			return v
		}
		return r
	}
	return v
}

func (c *Context) PostValueBoolSlice(name string) ([]bool, bool) {
	var r []bool
	q, ok := c.body[name]
	if !ok {
		r := make([]bool, len(q))
		for index, v := range q {
			t, err := strconv.ParseBool(v)
			if err != nil {
				return r, false
			}
			r[index] = t
		}
		return r, true
	}
	return r, false
}

func (c *Context) JSONValue(name string) (string, bool) {
	q, ok := c.json[name].(string)
	return q, ok
}

func (c *Context) JSONValueDefault(name string, v string) (string, bool) {
	q, ok := c.json[name].(string)
	if !ok {
		return v, ok
	}
	return q, ok
}

func (c *Context) JSONValueSlice(name string) ([]string, bool) {
	var r []string
	q, ok := c.json[name].([]interface{})
	if !ok {
		return nil, ok
	}
	r = make([]string, len(q))
	for index, v := range q {
		a, ok := v.(string)
		if !ok {
			return nil, false
		}
		r[index] = a
	}
	return r, ok
}

func (c *Context) JSONValueInt(name string) (int, bool) {
	r, ok := c.json[name].(float64)
	if ok {
		a := int(r)
		if float64(a) == r {
			return a, ok
		}
		return a, false
	}
	return 0, ok
}

func (c *Context) JSONValueIntDefault(name string, v int) int {
	q, ok := c.json[name].(float64)
	if ok {
		a := int(q)
		if float64(a) == q {
			return a
		}
		return v
	}
	return v
}

func (c *Context) JSONValueIntSlice(name string) ([]int, bool) {
	var r []int
	q, ok := c.json[name].([]interface{})
	if !ok {
		return nil, ok
	}
	r = make([]int, len(q))
	for index, v := range q {
		a, ok := v.(float64)
		if !ok {
			return nil, false
		}
		i := int(a)
		if float64(i) != a {
			return nil, false
		}
		r[index] = i
	}
	return r, ok
}

func (c *Context) JSONValueFloat64(name string) (float64, bool) {
	q, ok := c.json[name].(float64)
	return q, ok
}

func (c *Context) JSONValueFloat64Default(name string, v float64) float64 {
	q, ok := c.json[name].(float64)
	if !ok {
		return v
	}
	return q
}

func (c *Context) JSONValueFloat64Slice(name string) ([]float64, bool) {
	var r []float64
	q, ok := c.json[name].([]interface{})
	if !ok {
		return nil, ok
	}
	r = make([]float64, len(q))
	for index, v := range q {
		a, ok := v.(float64)
		if !ok {
			return nil, false
		}
		r[index] = a
	}
	return r, ok
}

func (c *Context) JSONValueBool(name string) (bool, bool) {
	q, ok := c.json[name].(bool)
	return q, ok
}

func (c *Context) JSONValueBoolDefault(name string, v bool) bool {
	q, ok := c.json[name].(bool)
	if !ok {
		return v
	}
	return q
}

func (c *Context) JSONValueBoolSlice(name string) ([]bool, bool) {
	var r []bool
	q, ok := c.json[name].([]interface{})
	if !ok {
		return nil, ok
	}
	r = make([]bool, len(q))
	for index, v := range q {
		b, ok := v.(bool)
		if !ok {
			return nil, false
		}
		r[index] = b
	}
	return r, ok
}

func (c *Context) JSONValueType(name string, ptr interface{}) bool {
	q, ok := c.json[name].(string)
	if !ok {
		return false
	}
	err := json.Unmarshal([]byte(q), ptr)
	return err == nil
}

func (c *Context) FilesUpload(name string, dist string, modifyName func(string) string) ([]string, error) {
	var names = make([]string, 0)
	files, ok := c.files[name]
	if !ok {
		return names, errors.New("the files is not exist")
	}
	if !strings.HasSuffix(dist, "/") {
		dist = dist + "/"
	}
	for _, file := range files {
		name := dist + modifyName(file.Filename)
		names = append(names, name)
		sf, err := file.Open()
		if err != nil {
			return names, err
		}
		f, err := os.Create(name)
		if err != nil {
			return names, err
		}
		_, err = io.Copy(f, sf)
		if err != nil {
			return names, err
		}
		sf.Close()
	}
	return names, nil
}

func (c *Context) File(name string) (*multipart.FileHeader, bool) {
	files, ok := c.files[name]
	if ok {
		return files[0], true
	}
	return nil, false
}

func (c *Context) Files(name string) ([]*multipart.FileHeader, bool) {
	f, ok := c.files[name]
	return f, ok
}
