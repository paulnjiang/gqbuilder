package gqbuilder

import (
	"regexp"
	"strings"
	"database/sql"
)

// Query is a SQL statement. it can be select, update, insert and delete
type Query struct {
	elements   []element
	builder    *Builder
	method     queryMethod
	flagOr     bool
	flagNot    bool
	isDistinct bool
}

func newQuery(bd *Builder) *Query {
	q := new(Query)
	q.builder = bd
	q.method = selectMethod
	return q
}

func (q *Query) clone() *Query {
	qq := *q
	return &qq
}

func (q *Query) addElement(elm element) bool {
	q.elements = append(q.elements, elm)
	return true
}

func (q *Query) replaceElement(elm element) int {
	n := 0
	for i := 0; i < len(q.elements); i++ {
		c := q.elements[i]
		if c.getElementName() == elm.getElementName() {
			q.elements[i] = elm
			n++
		}
	}
	return n
}

func (q *Query) clearElements(elementName string) *Query {
	p := 0
	for _, elm := range q.elements {
		if elm.getElementName() != elementName {
			q.elements[p] = elm
			p++
		}
	}
	q.elements = q.elements[:p]
	return q
}

func (q *Query) replaceOrAdd(elm element) int {
	var n int
	if n := q.replaceElement(elm); n == 0 {
		q.addElement(elm)
		return 1
	}
	return n
}

func (q *Query) getNot() bool {
	isNot := q.flagNot
	q.flagNot = false
	return isNot
}

func (q *Query) getOr() bool {
	isOr := q.flagOr
	q.flagOr = false
	return isOr
}

func (q *Query) getElement(name string) (element, bool) {
	for i := range q.elements {
		if q.elements[i].getElementName() == name {
			return q.elements[i], true
		}
	}
	return nil, false
}

func (q *Query) getElements(name string) ([]element, int) {
	// elms := make([]element, 0)
	var elms []element
	for _, elm := range q.elements {
		if elm.getElementName() == name {
			elms = append(elms, elm)
		}
	}
	return elms, len(elms)
}

func (q *Query) splitAlias(mixture string) (name, alias string) {
	re, _ := regexp.Compile(`\s+(as|AS|As|aS)\s+`)
	ns := re.Split(strings.TrimSpace(mixture), 2)
	if len(ns) == 2 {
		return ns[0], ns[1]
	}
	return ns[0], ""
}

// Not is NOT operator
func (q *Query) Not() *Query {
	q.flagNot = true
	return q
}

// Or is OR operator
func (q *Query) Or() *Query {
	q.flagOr = true
	return q
}

// From is FROM clause
func (q *Query) From(tables ...string) *Query {
	for _, tab := range tables {
		var cls fromClause
		nam, alias := q.splitAlias(tab)
		if alias != "" {
			cls.alias = alias
		}
		cls.elementName = "from"
		cls.tableName = nam
		q.addElement(cls)
	}
	return q
}

// Select add a SELECT clause to query statement
func (q *Query) Select(columns ...string) *Query {
	q.method = selectMethod
	if len(columns) == 0 {
		var cls columnClause
		cls.name = "*"
		cls.elementName = "column"
		q.addElement(cls)
		return q
	}
	for _, column := range columns {
		var cls columnClause
		nam, alias := q.splitAlias(column)
		if alias != "" {
			cls.alias = alias
		}
		cls.name = nam
		cls.elementName = "column"
		q.addElement(cls)
	}
	return q
}

// RawSelect add a raw expression to select clause
func (q *Query) RawSelect(expression string) *Query {
	var cls rawColumnClause
	cls.expression = expression
	cls.elementName = "RawColumn"
	q.addElement(cls)
	return q
}

// Distinct add DISTINCT clause to query
func (q *Query) Distinct() *Query {
	q.isDistinct = true
	return q
}

// Where add WHERE constraint to query
func (q *Query) Where(columnName string, sign string, value interface{}) *Query {
	// cls := new(compareCondition)
	var cls compareCondition
	cls.columnName = columnName
	cls.sign = sign
	cls.value = value
	cls.isNot = q.getNot()
	cls.isOr = q.getOr()
	cls.elementName = "where"
	q.addElement(cls)
	return q
}

func (q *Query) OrWhere(columnName string, sign string, value interface{}) *Query {
	return q.Or().Where(columnName, sign, value)
}

func (q *Query) WhereLike(columnName string, like string) *Query {
	var cls likeCondition
	cls.columnName = columnName
	cls.like = like
	cls.isNot = q.getNot()
	cls.isOr = q.getOr()
	cls.elementName = "where"
	q.addElement(cls)
	return q
}

func (q *Query) OrWhereLike(columnName string, like string) *Query {
	return q.Or().WhereLike(columnName, like)
}

func (q *Query) WhereNotLike(columnName string, like string) *Query {
	return q.Not().WhereLike(columnName, like)
}

func (q *Query) OrWhereNotLike(columnName string, like string) *Query {
	return q.Or().Not().WhereLike(columnName, like)
}

func (q *Query) Between(columnName string, from interface{}, to interface{}) *Query {
	// btw := new(betweenCondition)
	var btw betweenCondition
	btw.columnName = columnName
	btw.from = from
	btw.to = to
	btw.isNot = q.getNot()
	btw.isOr = q.getOr()
	btw.elementName = "where"
	q.addElement(btw)
	return q
}

func (q *Query) OrBetween(columnName string, from interface{}, to interface{}) *Query {
	return q.Or().Between(columnName, from, to)
}

func (q *Query) NotBetween(columnName string, from interface{}, to interface{}) *Query {
	return q.Not().Between(columnName, from, to)
}

func (q *Query) OrNotBetween(columnName string, from interface{}, to interface{}) *Query {
	return q.Or().Not().Between(columnName, from, to)
}

func (q *Query) WhereIn(columnName string, members ...interface{}) *Query {
	// in := new(inCondition)
	var in inCondition
	in.columnName = columnName
	in.members = members
	in.elementName = "where"
	in.isNot = q.getNot()
	in.isOr = q.getOr()
	q.addElement(in)
	return q
}

func (q *Query) WhereNotIn(columnName string, members ...interface{}) *Query {
	return q.Not().WhereIn(columnName, members...)
}

func (q *Query) OrWhereIn(columnName string, members ...interface{}) *Query {
	return q.Or().WhereIn(columnName, members...)
}

func (q *Query) OrWhereNotIn(columnName string, members ...interface{}) *Query {
	return q.Or().Not().WhereIn(columnName, members...)
}

func (q *Query) WhereNull(columnName string) *Query {
	// cls := new(NullCondition)
	var cls nullCondition
	cls.columnName = columnName
	cls.elementName = "where"
	cls.isNot = q.getNot()
	cls.isOr = q.getOr()
	q.addElement(cls)
	return q
}

func (q *Query) WhereNotNull(columnName string) *Query {
	return q.Not().WhereNull(columnName)
}

func (q *Query) OrWhereNull(columnName string) *Query {
	return q.Or().WhereNull(columnName)
}

func (q *Query) OrWhereNotNull(columnName string) *Query {
	return q.Or().Not().WhereNull(columnName)
}

/*
func (q *Query) WhereTrue(columnName string) *Query {
	// cls := new(booleanCondition)
	var cls booleanCondition
	cls.columnName = columnName
	cls.value = true
	cls.elementName = "where"
	cls.isNot = q.getNot()
	cls.isOr = q.getOr()
	q.addElement(cls)
	return q
}

func (q *Query) OrWhereTrue(columnName string) *Query {
	return q.Or().WhereTrue(columnName)
}

func (q *Query) WhereFalse(columnName string) *Query {
	// cls := new(booleanCondition)
	var cls booleanCondition
	cls.columnName = columnName
	cls.value = false
	cls.elementName = "where"
	cls.isNot = q.getNot()
	cls.isOr = q.getOr()
	q.addElement(cls)
	return q
}

func (q *Query) OrWhereFalse(columnName string) *Query {
	return q.Or().WhereFalse(columnName)
}*/

// WhereInQuery add a sub query to query
func (q *Query) WhereInQuery(columnName string, subQuery *Query) *Query {
	var cls inQueryCondition
	cls.columnName = columnName
	cls.subQuery = subQuery
	cls.elementName = "where"
	cls.isNot = q.getNot()
	cls.isOr = q.getOr()
	q.addElement(cls)
	return q
}

func (q *Query) OrWhereInQuery(columnName string, subQuery *Query) *Query {
	return q.Or().WhereInQuery(columnName, subQuery)
}

func (q *Query) WhereNotInQuery(columnName string, subQuery *Query) *Query {
	return q.Not().WhereInQuery(columnName, subQuery)
}

func (q *Query) OrWhereNotInQuery(columnName string, subQuery *Query) *Query {
	return q.Or().Not().WhereInQuery(columnName, subQuery)
}

func (q *Query) WhereExists(subQuery *Query) *Query {
	var cls existsCondition
	cls.subQuery = subQuery
	cls.elementName = "where"
	cls.isNot = q.getNot()
	cls.isOr = q.getOr()
	q.addElement(cls)
	return q
}

func (q *Query) WhereNotExists(subQuery *Query) *Query {
	return q.Not().WhereExists(subQuery)
}

func (q *Query) join(typ joinType, tableName, leftTable, sign, rightTable string) *Query {
	var cls joinClause
	cls.joinTyp = typ
	cls.left = leftTable
	cls.right = rightTable
	cls.table = tableName
	cls.sign = sign
	cls.elementName = "join"
	q.addElement(cls)
	return q
}

// LeftJoin add a left join clause
func (q *Query) LeftJoin(tableName, leftColumn, sign, rightColumn string) *Query {
	return q.join(leftJoin, tableName, leftColumn, sign, rightColumn)
}

// RightJoin add a right join clause
func (q *Query) RightJoin(tableName, leftColumn, sign, rightColumn string) *Query {
	return q.join(rightJoin, tableName, leftColumn, sign, rightColumn)
}

// Join add a inner join clause
func (q *Query) Join(tableName, leftColumn, sign, rightColumn string) *Query {
	return q.join(innerJoin, tableName, leftColumn, sign, rightColumn)
}

// OrderBy add ORDER BY clause to query
func (q *Query) OrderBy(columnName string) *Query {
	// cls := new(orderByClause)
	var cls orderByClause
	cls.columnName = columnName
	cls.elementName = "order"
	q.addElement(cls)
	return q
}

// OrderByDesc add ORDER BY DESC clause to query
func (q *Query) OrderByDesc(columnName string) *Query {
	// cls := new(orderByClause)
	var cls orderByClause
	cls.columnName = columnName
	cls.desc = true
	cls.elementName = "order"
	q.addElement(cls)
	return q
}

// GroupBy add GROUP BY clause to query
func (q *Query) GroupBy(columnNames ...string) *Query {
	// cls := new(groupByClause)
	var cls groupByClause
	cls.columnNames = columnNames
	cls.elementName = "group"
	q.replaceOrAdd(cls)
	return q
}

// Having add Having clause to query
func (q *Query) Having(columnName string, sign string, value interface{}) *Query {
	var cls compareCondition
	cls.columnName = columnName
	cls.sign = sign
	cls.value = value
	cls.elementName = "having"
	cls.isNot = q.getNot()
	cls.isOr = q.getOr()
	q.addElement(cls)
	return q
}

func (q *Query) OrHaving(columnName string, sign string, value interface{}) *Query {
	return q.Or().Having(columnName, sign, value)
}

func (q *Query) HavingRaw(expression string) *Query {
	var cls rawCodition
	cls.expression = expression
	cls.elementName = "having"
	cls.isNot = q.getNot()
	cls.isOr = q.getOr()
	q.addElement(cls)
	return q
}

func (q *Query) OrHavingRaw(expression string) *Query {
	return q.Or().HavingRaw(expression)
}

// Limit add LIMIT clause to query
func (q *Query) Limit(rowCount int) *Query {
	// cls := new(limitClause)
	var cls limitClause
	if rowCount < 0 {
		cls.rowCount = 0
	} else {
		cls.rowCount = int(rowCount)
	}
	cls.elementName = "limit"
	q.replaceOrAdd(cls)
	return q
}

// Offset add OFFSET clause to query
func (q *Query) Offset(rows int) *Query {
	// cls := new(offsetClause)
	var cls offsetClause
	if rows < 0 {
		cls.offset = 0
	} else {
		cls.offset = int(rows)
	}
	cls.elementName = "offset"
	q.replaceOrAdd(cls)
	return q
}

// Insert build a insert statement 
func (q *Query) Insert(columns []string, values []interface{}) *Query {
	q.method = insertMethod
	q.clearElements("insert")
	var cls insertClause
	cls.columns = columns
	cls.values = values
	cls.elementName = "insert"
	q.addElement(cls)
	return q
}

func (q *Query) InsertFromQuery(subq *Query) *Query {
	q.method = insertMethod
	q.clearElements("insert")
	var cls insertClause
	cls.elementName = "insert"
	cls.subQuery = subq
	q.addElement(cls)
	return q
}

func (q *Query) InsertFromMap(item map[string]interface{}) *Query {
	q.method = insertMethod
	q.clearElements("insert")
	var cls insertClause
	cls.elementName = "insert"
	for k, v := range item {
		cls.columns = append(cls.columns, k)
		cls.values = append(cls.values, v)
	}
	q.addElement(cls)
	return q
}

// TODO: insert data from structs
// func (q *Query) InsertObject(objs struct{}) *Query {
// }

// Update build a update statement
func (q *Query) Update(item map[string]interface{}) *Query {
	q.clearElements("update")
	q.method = updateMethod
	var cls updateClause
	cls.item = item
	cls.elementName = "update"
	q.addElement(cls)
	return q
}

// TODO: update data from structs
// func (q *Query) updateObject(obj struct{}) *Query {
// }

// Delete build a delete statement
func (q *Query) Delete() *Query {
	q.method = deleteMethod
	return q
}

// ToString replace all placeholders to value in sql statement, and a error will be return when sql variable can't be
// convert to string
func (q *Query) ToString() (string, error) {
	cmpl := q.builder.cmpl.clone()
	rst, err := cmpl.compile(q)
	if err != nil {
		return "", err
	}
	return rst.ToString()
}

// ToPrepared return a string with placeholders, and a variables list
func (q *Query) ToPrepared() (string, []interface{}, error) {
//	var rst *SQLResult
	cmpl := q.builder.cmpl.clone()
	rst, err := cmpl.compile(q)
	if err != nil {
		return "", nil, err
	}
	sql, args := rst.ToPrepared()
	return sql, args, nil
}

// Do execute the query with DB.QueryRow() 
func (q *Query) Do() (*sql.Row, error) {
	sql, values, err := q.ToPrepared()
	if err != nil {
		return nil, err
	}
	return q.builder.pool.QueryRow(sql, values...), nil
}

// Get execute the query with DB.Query()
func (q *Query) Get() (*sql.Rows, error) {
	sql, values, err := q.ToPrepared()
	if err != nil {
		return nil, err
	}
	return q.builder.pool.Query(sql, values...)
}
