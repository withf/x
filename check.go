package x

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

type CheckExist struct {
	Name       string
	From       string
	Fail       interface{}
	StatusCode int
}

type CheckInt struct {
	Name        string
	From        string
	Fail        interface{}
	Max         interface{}
	Min         interface{}
	MaxFail     interface{}
	MinFail     interface{}
	StatusCode  int
	NotRequired bool
}

type CheckFloat64 struct {
	Name        string
	From        string
	Fail        interface{}
	Max         interface{}
	Min         interface{}
	MaxFail     interface{}
	MinFail     interface{}
	StatusCode  int
	NotRequired bool
}

type CheckString struct {
	Name          string
	From          string
	Fail          interface{}
	StatusCode    int
	MaxLength     interface{}
	MinLength     interface{}
	Blank         bool
	MaxLengthFail interface{}
	MinLengthFail interface{}
	BlankFail     interface{}
	NotTrim       bool
	NotRequired   bool
}

type CheckBool struct {
	Name        string
	From        string
	Fail        interface{}
	StatusCode  int
	NotRequired bool
}

type CheckFile struct {
	Name        string
	Fail        interface{}
	Max         interface{}
	Min         interface{}
	MaxFail     interface{}
	MinFail     interface{}
	StatusCode  int
	NotRequired bool
}

func Check(vs ...interface{}) *group {
	g := group{
		path:   "",
		before: []contextFunction{},
		after:  []contextFunction{},
	}
	if len(vs) == 0 {
		return &g
	}
	f := checkType(vs)
	g.before = append(g.before, f...)
	return &g
}

func checkType(vs []interface{}) []contextFunction {
	functions := make([]contextFunction, len(vs))
	for index, v := range vs {
		switch t := v.(type) {
		case CheckInt:
			functions[index] = checkInt(t)
		case CheckBool:
			functions[index] = checkBool(t)
		case CheckString:
			functions[index] = checkString(t)
		case CheckFloat64:
			functions[index] = checkFloat64(t)
		case CheckExist:
			functions[index] = checkExist(t)
		case CheckFile:
			functions[index] = checkFile(t)
		default:
			a, ok := v.(func(Context))
			if ok {
				functions[index] = a
			} else {
				panic("The type must be CheckInt/CheckFloat64/CheckBool/CheckString/CheckExist/CheckFile/Context Function")
			}
		}
	}
	return functions
}

func checkInt(v CheckInt) contextFunction {
	if v.StatusCode == 0 {
		v.StatusCode = HTTP_BAD_REQUEST
	}
	from := strings.ToLower(v.From)
	if from != "query" && from != "body" && from != "json" {
		panic("The check type attribute 'From' must be 'query'/'body'/'json'")
	}
	return func(c Context) {
		var ok bool
		var r int
		var err error
		if from == "query" {
			r, ok, err = c.queryInt(v.Name)
		} else if from == "body" {
			r, ok, err = c.postValueInt(v.Name)
		} else {
			r, ok = c.JSONValueInt(v.Name)
		}
		if !ok {
			if v.NotRequired && err == nil {
				c.Next()
				return
			}
			if v.Fail == nil {
				c.WriteString("the param "+v.Name+" is required and must int", v.StatusCode)
				return
			} else {
				switch s := v.Fail.(type) {
				case string:
					c.WriteString(s, v.StatusCode)
					return
				default:
					c.SendJSON(s, v.StatusCode)
					return
				}
			}
		}
		max, ok := v.Max.(int)
		if ok {
			if r > max {
				if v.MaxFail != nil {
					switch s := v.MaxFail.(type) {
					case string:
						c.WriteString(s, v.StatusCode)
						return
					default:
						c.SendJSON(s, v.StatusCode)
						return
					}
				} else {
					c.WriteString("the param "+v.Name+" is must < "+strconv.Itoa(max), v.StatusCode)
					return
				}
			}
		}
		min, ok := v.Min.(int)
		if ok {
			if r < min {
				if v.MinFail != nil {
					switch s := v.MinFail.(type) {
					case string:
						c.WriteString(s, v.StatusCode)
						return
					default:
						c.SendJSON(s, v.StatusCode)
						return
					}
				} else {
					c.WriteString("the param "+v.Name+" is must > "+strconv.Itoa(max), v.StatusCode)
					return
				}
			}
		}
		c.Next()
	}
}

func checkBool(v CheckBool) contextFunction {
	if v.StatusCode == 0 {
		v.StatusCode = HTTP_BAD_REQUEST
	}
	from := strings.ToLower(v.From)
	if from != "query" && from != "body" && from != "json" {
		panic("The check type attribute 'From' must be 'query'/'body'/'json'")
	}
	return func(c Context) {
		var ok bool
		var err error
		if from == "query" {
			_, ok, err = c.queryBool(v.Name)
		} else if from == "body" {
			_, ok, err = c.postValueBool(v.Name)
		} else {
			_, ok = c.JSONValueBool(v.Name)
		}
		if !ok {
			if v.NotRequired && err == nil {
				c.Next()
				return
			}
			if v.Fail == nil {
				c.WriteString("the param "+v.Name+" is required and must bool", v.StatusCode)
				return
			} else {
				switch s := v.Fail.(type) {
				case string:
					c.WriteString(s, v.StatusCode)
					return
				default:
					c.SendJSON(s, v.StatusCode)
					return
				}
			}
		}
		c.Next()
	}
}

func checkString(v CheckString) contextFunction {
	if v.StatusCode == 0 {
		v.StatusCode = HTTP_BAD_REQUEST
	}
	from := strings.ToLower(v.From)
	if from != "query" && from != "body" && from != "json" {
		panic("The check type attribute 'From' must be 'query'/'body'/'json'")
	}
	return func(c Context) {
		var ok bool
		var r string
		if from == "query" {
			r, ok = c.QueryString(v.Name)
		} else if from == "body" {
			r, ok = c.PostValue(v.Name)
		} else {
			r, ok = c.JSONValue(v.Name)
		}
		if !v.NotTrim {
			r = strings.Trim(r, " ")
		}
		if !ok {
			if v.NotRequired {
				c.Next()
				return
			}
			if v.Fail == nil {
				c.WriteString("the param "+v.Name+" is required.", v.StatusCode)
				return
			} else {
				switch s := v.Fail.(type) {
				case string:
					c.WriteString(s, v.StatusCode)
					return
				default:
					c.SendJSON(s, v.StatusCode)
					return
				}
			}
		}
		if !v.Blank && r == "" {
			if v.BlankFail != nil {
				switch s := v.BlankFail.(type) {
				case string:
					c.WriteString(s, v.StatusCode)
					return
				default:
					c.SendJSON(s, v.StatusCode)
					return
				}
			} else {
				c.WriteString("The param "+v.Name+" is not blank.", v.StatusCode)
				return
			}
		}
		strLen := utf8.RuneCountInString(r)
		max, ok := v.MaxLength.(int)
		if ok {
			if strLen > max {
				if v.MaxLengthFail != nil {
					switch s := v.MaxLengthFail.(type) {
					case string:
						c.WriteString(s, v.StatusCode)
						return
					default:
						c.SendJSON(s, v.StatusCode)
						return
					}
				} else {

					c.WriteString("the param "+v.Name+" length is must < "+strconv.Itoa(max), v.StatusCode)
					return
				}
			}
		}
		min, ok := v.MinLength.(int)
		if ok {
			if strLen < min {
				if v.MinLengthFail != nil {
					switch s := v.MinLengthFail.(type) {
					case string:
						c.WriteString(s, v.StatusCode)
						return
					default:
						c.SendJSON(s, v.StatusCode)
						return
					}
				} else {
					c.WriteString("the param "+v.Name+" length is must > "+strconv.Itoa(min), v.StatusCode)
					return
				}
			}
		}
		c.Next()
	}
}

func checkFloat64(v CheckFloat64) contextFunction {
	if v.StatusCode == 0 {
		v.StatusCode = HTTP_BAD_REQUEST
	}
	from := strings.ToLower(v.From)
	if from != "query" && from != "body" && from != "json" {
		panic("The check type attribute 'From' must be 'query'/'body'/'json'")
	}
	return func(c Context) {
		var ok bool
		var r float64
		var err error
		if from == "query" {
			r, ok, err = c.queryFloat64(v.Name)
		} else if from == "body" {
			r, ok, err = c.postValueFloat64(v.Name)
		} else {
			r, ok = c.JSONValueFloat64(v.Name)
		}
		if !ok {
			if v.NotRequired && err == nil {
				c.Next()
				return
			}
			if v.Fail == nil {
				c.WriteString("the param "+v.Name+" is required.", v.StatusCode)
				return
			} else {
				switch s := v.Fail.(type) {
				case string:
					c.WriteString(s, v.StatusCode)
					return
				default:
					c.SendJSON(s, v.StatusCode)
					return
				}
			}
		}
		max, ok := v.Max.(float64)
		if ok {
			if r > max {
				if v.MaxFail != nil {
					switch s := v.MaxFail.(type) {
					case string:
						c.WriteString(s, v.StatusCode)
						return
					default:
						c.SendJSON(s, v.StatusCode)
						return
					}
				} else {

					c.WriteString("the param "+v.Name+" is must < "+strconv.FormatFloat(max, 'E', -1, 64), v.StatusCode)
					return
				}
			}
		}
		min, ok := v.Min.(float64)
		if ok {
			if r < min {
				if v.MinFail != nil {
					switch s := v.MinFail.(type) {
					case string:
						c.WriteString(s, v.StatusCode)
						return
					default:
						c.SendJSON(s, v.StatusCode)
						return
					}
				} else {
					c.WriteString("the param "+v.Name+" is must > "+strconv.FormatFloat(min, 'E', -1, 64), v.StatusCode)
					return
				}
			}
		}
		c.Next()
	}
}

func checkExist(v CheckExist) contextFunction {
	if v.StatusCode == 0 {
		v.StatusCode = HTTP_BAD_REQUEST
	}
	from := strings.ToLower(v.From)
	if from != "query" && from != "body" && from != "json" {
		panic("The check type attribute 'From' must be 'query'/'body'/'json'")
	}
	return func(c Context) {
		var ok bool
		if from == "query" {
			_, ok = c.query[v.Name]
		} else if from == "form" {
			_, ok = c.body[v.Name]
		} else {
			_, ok = c.json[v.Name]
		}
		if !ok {
			if v.Fail == nil {
				c.WriteString("the param "+v.Name+" is required.", v.StatusCode)
				return
			} else {
				switch s := v.Fail.(type) {
				case string:
					c.WriteString(s, v.StatusCode)
					return
				default:
					c.SendJSON(s, v.StatusCode)
					return
				}
			}
		}
		c.Next()
	}
}

func checkFile(v CheckFile) contextFunction {
	if v.StatusCode == 0 {
		v.StatusCode = HTTP_BAD_REQUEST
	}
	return func(c Context) {
		files, ok := c.files[v.Name]
		if !ok {
			if v.NotRequired {
				c.Next()
				return
			}
			if v.Fail == nil {
				c.WriteString("the param "+v.Name+" is required.", v.StatusCode)
				return
			} else {
				switch s := v.Fail.(type) {
				case string:
					c.WriteString(s, v.StatusCode)
					return
				default:
					c.SendJSON(s, v.StatusCode)
					return
				}
			}
		}
		r := len(files)
		max, ok := v.Max.(int)
		if ok {
			if r > max {
				if v.MaxFail != nil {
					switch s := v.MaxFail.(type) {
					case string:
						c.WriteString(s, v.StatusCode)
						return
					default:
						c.SendJSON(s, v.StatusCode)
						return
					}
				} else {
					c.WriteString("the file param "+v.Name+" is must < "+strconv.Itoa(max), v.StatusCode)
					return
				}
			}
		}
		min, ok := v.Min.(int)
		if ok {
			if r < min {
				if v.MinFail != nil {
					switch s := v.MinFail.(type) {
					case string:
						c.WriteString(s, v.StatusCode)
						return
					default:
						c.SendJSON(s, v.StatusCode)
						return
					}
				} else {
					c.WriteString("the file param "+v.Name+" is must > "+strconv.Itoa(max), v.StatusCode)
					return
				}
			}
		}
		c.Next()
	}
}
