/*
SQL builder implement by Golang
*/
package gqbuilder

import (
	"fmt"
	"testing"
	// "reflect"
)

func simpleQuery() {
	db := NewBuilder(MySQL)
	q := db.Query("user")
	q.Select("name", "age").
		Join("class", "class.id", "=", "student.cid").
		LeftJoin("school", "school.id", "!=", "class.sid").
		Where("age", ">", 15).
		Where("name", "=", "lisi").
		WhereFalse("class").
		WhereTrue("hasName").
		WhereNotIn("fuck", []interface{}{1, 2, 3, 4}).
		OrWhere("sss", "<=", 123).
		GroupBy("name", "age").
		OrderBy("sss").
		OrderByDesc("class").
		Limit(10).
		Offset(30)
	// sqlres := newSQLResult(PlaceHolder, ":")
	subq := q.Clone()
	subq.Select("name").From("user").WhereInQuery("class", q)
	c := newMySQLCompiler()
	// t := reflect.TypeOf(c).String()
	// v := reflect.TypeOf(c).NumMethod()
	res := c.compile(subq)
	str, _ := res.ToString()
	fmt.Printf("%v\n", str)
}

func TestQuery(t *testing.T) {
	simpleQuery()
}
