package gqbuilder


type mysqlCompiler struct {
	baseCompiler
}

func newMySQLCompiler() *mysqlCompiler {
	c := new(mysqlCompiler)
	c.paramsPattern = PlaceHolder
	c.symbolPrefix = "?"
	c.leftIdentifier = "`"
	c.righIdentifier = "`"
	return c
}
