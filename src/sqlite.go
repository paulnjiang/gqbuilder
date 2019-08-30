package gqbuilder

type sqliteCompiler struct {
	baseCompiler
}

func newSQLiteCompiler() *sqliteCompiler {
	c := new(sqliteCompiler)
	c.paramsPattern = PlaceHolder
	c.symbolPrefix = "?"
	c.leftIdentifier = "\""
	c.righIdentifier = "\""
	return c
}
