# Golang SQL Builder

Golang SQL builder(package: gqbuilder) is written by Golang, and is a simple and easy to use SQL statement builder.


# Examples

## Setup database
```go
import (
    "gqbuilder"
    "database/sql"
    "fmt"
)

db := *sql.DB
bdr := NewBuilder(gqbuilder.MySQL, db)
```

## Create a query
```go
q := bdr.Query("user")
```

## Add conditions into query
```go
q.Select("id", "name", "age", "telphone as phone").Where("age", ">", 20).OrWhere("las_login", "!=", time.now())
```

## Join
```go
q.Select("id", "name", "age", "telphone as phone").LeftJoin("address", "user.id", "=", "address.uid")
```

## Having 
```go
q.Select("id", "name", "age", "telphone as phone").GroupBy("age").Having("id", "<", 100)
```

## Sub query
```go
qq := bdr.Query("address").Select("uid").WhereNotNull("home address")

q := bdr.Query("user").Select("id", "name", "age", "telphone as phone").Where("id", ">", "14").WhereInQuery("id", qq)
```

## Insert
```go
q := bdr.Query("user").Insert([]string{"name", "age"}, []interface{}{"bob", 18})
q2 := bdr.Query("user").InsertFromMap(map[string]interface{}{"name": "bob", "age": 19})

qq := bdr.Query("other").Where("id", "<", 100)
q3 := bdr.Query("user").InsertFromQuery(qq)
```

## Update
```go
q := bdr.Query("user").Update(map[string]interface{}{"name": "bob", "age": 19}).Where("id", "=", 119)
```

## Delete
```go
bdr.Query("user").Delete().Where("id", "=", 112)

qq := bdr.Query("other").Select("name").Where("class", "=", "9")
q := bdr.Query("user").Delete().WhereNotExists(qq)
```


# Result

## ToString()

Return a full sql statement

```go
q := bdr.Query("user").Select("id", "name", "age", "telphone as phone").OrderBy("age")
sql, err := q.ToString()

fmt.Println(ssql)
```
```sql
SELECT `id`, `name`, `age`, `telphone` AS `phone` FROM `user` ORDER BY `age` ASC 
```

## ToPrepared()

Return a sql statement with placeholders, and a []interface{} include values

```go
q := bdr.Query("user")
q.Select("id", "name", "age", "telphone as phone").Between("age", 10, 15).NotBetween("phone", 139, 189)

raw, args, err := q.ToPrepared()
fmt.Printf("%#v\n%#v\n", raw, args)
```

```sql
"SELECT `id`, `name`, `age`, `telphone` AS `phone` FROM `user` WHERE `age` BETWEEN ? AND ? AND `phone` NOT BETWEEN ? AND ?"
[]interface {}{10, 15, 139, 189}
```

## Do()

DO() is a shortcut which execute insert, update, delete on database. 

It call *sql.DB.QueryRow(), and return *sql.Row.


## Get()

Get() is shortcut which execute query on database.

It call *sql.DB.Query(), and return *sql.Rows.