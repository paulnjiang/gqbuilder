/*
SQL builder implement by Golang
*/
package gqbuilder

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

const dbtype = MySQL


func TestSelect(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("tableA as t1")
	// q.Select("id", "name as zhangsan", "age", "telphone as phone")
	q.Where("t1.age", "<", 10)
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test select error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test select: %s \n", ssql)
}

func TestRawSelect(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "count(*)")
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test raw select error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test raw select: %s\n", ssql)
}

func TestDistinct(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age").Distinct()
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test distinct error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test distinct: %s\n", ssql)
}

func TestWhere(t *testing.T) {
	var ssql string
	var e error
	now := time.Now()
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "telphone as phone").Where("age", ">", now).OrWhere("id", "!=", 1)
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test where error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test where: %s\n", ssql)
}

func TestLike(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "telphone as phone").WhereLike("name", "%abc%")
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test like error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test like: %s\n", ssql)
}

func TestBetween(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "telphone as phone").Between("age", 10, 15).NotBetween("phone", 139, 189)
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test between error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test between: %s \n", ssql)
	raw, args, _ := q.ToPrepared()
	fmt.Printf("%#v\n%#v\n", raw, args)
}

func TestWhereIn(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "telphone as phone").WhereIn("name", "lily", "abby", "candy").OrWhereNotIn("id", 13, 12, 12, 22)
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test in error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test in: %s \n", ssql)
	raw, args, _ := q.ToPrepared()
	fmt.Printf("%#v\n%#v\n", raw, args)
}

func TestNull(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "telphone as phone").WhereNull("position").OrWhereNotNull("status")
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test null error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test null: %s\n", ssql)
}
/*
func TestBoolean(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("tableA")
	q.Select("id", "name", "age", "telphone as phone").WhereFalse("married").OrWhereTrue("retire")
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test boolean error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test boolean: %s\n", ssql)
}*/


func TestOrder(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user").Select("id", "name", "age", "telphone as phone").OrderBy("age")
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test order by error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test order by: %s \n", ssql)
}

func TestGroup(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "telphone as phone").GroupBy("age").OrderBy("age")
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test group error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test group: %s \n", ssql)
}

func TestHaving(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "telphone as phone").GroupBy("age").Having("id", "<", 100)
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test having error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test having: %s \n", ssql)
}

func TestLimitOffset(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "telphone as phone").Limit(10).Offset(5)
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test limit error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test limit: %s \n", ssql)
}

func TestInQuery(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user").Select("id", "name", "age", "telphone as phone").Where("id", ">", "14")
	qq := bdr.Query("address").Select("id").WhereNotNull("home address").Where("NO", "=", 1000)
	q.WhereInQuery("id", qq)
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test in query error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test in query: %s \n", ssql)
	raw, args, _ := q.ToPrepared()
	fmt.Printf("%#v\n%#v\n", raw, args)
}

func TestExists(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "telphone as phone")
	qq := bdr.Query("addres").Where("id", "=", 100)
	q.WhereExists(qq)
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test exists error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test exists: %s \n", ssql)
}

func TestJoin(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user")
	q.Select("id", "name", "age", "telphone as phone")
	q.LeftJoin("address", "user.id", "=", "address.id")
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test join error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test join: %s \n", ssql)
}

func TestInsert(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	q := bdr.Query("user").Insert([]string{"name", "age"}, []interface{}{"bob", 18})
	q2 := bdr.Query("user").InsertFromMap(map[string]interface{}{"name": "bob", "age": 19})
	qq := bdr.Query("tableB").Where("id", "<", 100)
	q3 := bdr.Query("user").InsertFromQuery(qq)
	
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test insert error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test insert: %s \n", ssql)
	if ssql, e = q2.ToString(); e != nil {
		t.Errorf("test insert error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test insert: %s \n", ssql)
	if ssql, e = q3.ToString(); e != nil {
		t.Errorf("test insert error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test insert: %s \n", ssql)
}

func TestUpdate(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)	
	q := bdr.Query("user").Update(map[string]interface{}{"name": "bob", "age": 19}).Where("id", "=", 119)
	
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test insert error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test insert: %s \n", ssql)
}

func TestDelete(t *testing.T) {
	var ssql string
	var e error
	var con *sql.DB 
	bdr := NewBuilder(dbtype, con)
	qq := bdr.Query("tableB").Select("name").Where("class", "=", "9")
	q := bdr.Query("user").Delete().WhereNotExists(qq)
	
	if ssql, e = q.ToString(); e != nil {
		t.Errorf("test insert error: sql: %s error: %s\n", ssql, e)
		return
	}
	fmt.Printf("test insert: %s \n", ssql)
}