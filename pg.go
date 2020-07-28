package gqbuilder

type pgCompiler struct {
	baseCompiler
}

func newPostgreSQLCompiler() *pgCompiler {
	c := new(pgCompiler)
	c.paramsPattern = Ordinal
	c.symbolPrefix = "$"
	c.leftIdentifier = "\""
	c.righIdentifier = "\""
	return c
}
