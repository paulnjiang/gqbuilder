package gqbuilder

import (
	"database/sql"
)

// Builder is a type
type Builder struct {
	driver databaseType
	pool   *sql.DB
	cmpl   compiler
}

// NewBuilder return a Builder that had saved type of database driver
func NewBuilder(driver databaseType, db *sql.DB) *Builder {
	bdr := new(Builder)
	bdr.driver = driver
	bdr.pool = db
	bdr.cmpl = compilerFactory(driver)
	return bdr
}

// Query create a querier, which bound on a table
func (b *Builder) Query(tableName string) *Query {
	q := newQuery(b)
	q.From(tableName)
	return q
}
