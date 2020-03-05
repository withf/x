package x

type group struct {
	path   string
	before []contextFunction
	after  []contextFunction
	check bool
}

func Group(path string) *group {
	return &group{path: _singleSlash(path), before: []contextFunction{}, after: []contextFunction{}}
}

func (g *group) Group(path string, f ...contextFunction) *group {
	return &group{
		path:   combineSlash(g.path, path),
		before: append(g.before, f...),
		after:  g.after,
	}
}

func (g *group) Register(method, path string, f ...contextFunction) {
	gf := append(append(g.before, f...), g.after...)
	var r = &router{
		method:    method,
		path:      combineSlash(g.path, path),
		functions: append(append(globalBefore, gf...), globalAfter...),
	}
	routerMap[toUpper(method)][r.path] = r
}

func (g *group) Get(path string, f ...contextFunction) {
	g.Register(_GET, path, f...)
}

func (g *group) Post(path string, f ...contextFunction) {
	g.Register(_POST, path, f...)
}

func (g *group) Put(path string, f ...contextFunction) {
	g.Register(_PUT, path, f...)
}

func (g *group) Patch(path string, f ...contextFunction) {
	g.Register(_PATCH, path, f...)
}

func (g *group) DELETE(path string, f ...contextFunction) {
	g.Register(_DELETE, path, f...)
}

func (g *group) Use(f ...contextFunction) {
	g.before = append(g.before, f...)
}

func (g *group) After(f ...contextFunction) {
	g.after = append(g.after, f...)
}

func (g *group) Check(vs ...interface{}) *group {
	ag := g
	if !g.check {
		ag = &group{
			path:   g.path,
			before: g.before,
			after: g.after,
			check:  true,
		}
	}
	if len(vs) == 0 {
		return ag
	}
	f := checkType(vs)
	ag.before = append(ag.before, f...)
	return ag
}