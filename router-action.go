package x

func Register(method, path string, f ...contextFunction) {
	err := checkMethod(method)
	if err != nil {
		panic(err)
	}
	routerMap[toUpper(method)][path] = &router{
		method:    method,
		path:      _singleSlash(path),
		functions: append(append(globalBefore, f...), globalAfter...),
	}
}

func Get(path string, f ...contextFunction) {
	Register(_GET, path, f...)
}

func Post(path string, f ...contextFunction) {
	Register(_POST, path, f...)
}

func Put(path string, f ...contextFunction) {
	Register(_PUT, path, f...)
}

func Delete(path string, f ...contextFunction) {
	Register(_DELETE, path, f...)
}

func Patch(path string, f ...contextFunction) {
	Register(_PATCH, path, f...)
}

