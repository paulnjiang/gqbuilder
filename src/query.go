package gqbuilder

import (
	"regexp"
	"strings"
)

// Builder is a type
type Builder struct {
	engine databaseType
}

// NewBuilder return a Builder that had saved type of database engine
func NewBuilder(engine databaseType) *Builder {
	bd := new(Builder)
	bd.engine = engine
	return bd
}

// Query create a query on the table
func (b *Builder) Query(tableName string) *query {
	q := newQuery(b.engine)
	q.From(tableName)
	return q
}

type query struct {
	components []component
	engine     databaseType
	method     string
	flagOr     bool
	flagNot    bool
	isDistinct bool
}

func newQuery(engine databaseType) *query {
	q := new(query)
	q.engine = engine
	return q
}

func (q query) Clone() *query {
	qq := q
	qq.components = []component{}
	qq.flagNot = false
	qq.flagOr = false
	qq.isDistinct = false
	return &qq
}

func (q *query) Not() *query {
	q.flagNot = true
	return q
}

func (q *query) Or() *query {
	q.flagOr = true
	return q
}

func (q *query) From(tableName string) *query {
	var from fromClause
	from.componentName = "from"
	from.tableName = strings.TrimSpace(tableName)
	q.replaceOrAdd(from)
	return q
}

// Select select columns
func (q *query) Select(columns ...string) *query {
	if len(columns) == 0 {
		var cla columnClause
		cla.name = "*"
		cla.componentName = "column"
		q.addComponent(cla)
	} else {
		for _, column := range columns {
			if ok, _ := regexp.MatchString(`^\w+\s+(AS|as)\s+\w+$`, column); !ok {
				// cla := new(columnClause)
				var cla columnClause
				cla.name = column
				cla.componentName = "column"
				q.addComponent(cla)
			} else {
				sss := regexp.MustCompile(`\s+`).Split(column, 3)
				// cla := new(columnClause)
				var cla columnClause
				cla.name = sss[0]
				cla.alias = sss[2]
				cla.componentName = "column"
				q.addComponent(cla)
			}
		}
	}
	q.method = "select"
	return q
}

func (q *query) RawSelect(expression string) *query {
	var cla rawColumnClause
	cla.expression = expression
	cla.componentName = "column"
	q.addComponent(cla)
	return q
}

func (q *query) Distinct() *query {
	q.isDistinct = true
	return q
}

func (q *query) Where(columnName string, sign string, value interface{}) *query {
	// cla := new(compareCondition)
	var cla compareCondition
	cla.columnName = columnName
	cla.sign = sign
	cla.value = value
	cla.isNot = q.getNot()
	cla.isOr = q.getOr()
	cla.componentName = "where"
	q.addComponent(cla)
	return q
}

func (q *query) OrWhere(columnName string, sign string, value interface{}) *query {
	return q.Or().Where(columnName, sign, value)
}

func (q *query) OrWhereNot(columnName string, sign string, value interface{}) *query {
	return q.Or().Not().Where(columnName, sign, value)
}

func (q *query) Between(columnName string, from interface{}, to interface{}) *query {
	// btw := new(betweenCondition)
	var btw betweenCondition
	btw.columnName = columnName
	btw.from = from
	btw.to = to
	btw.isNot = q.getNot()
	btw.isOr = q.getOr()
	btw.componentName = "where"
	q.addComponent(btw)
	return q
}

func (q *query) OrBetween(columnName string, from interface{}, to interface{}) *query {
	return q.Or().Between(columnName, from, to)
}

func (q *query) NotBetween(columnName string, from interface{}, to interface{}) *query {
	return q.Not().Between(columnName, from, to)
}

func (q *query) OrNotBetween(columnName string, from interface{}, to interface{}) *query {
	return q.Or().Not().Between(columnName, from, to)
}

func (q *query) WhereIn(columnName string, members ...interface{}) *query {
	// in := new(inCondition)
	var in inCondition
	in.columnName = columnName
	in.members = members
	in.componentName = "where"
	in.isNot = q.getNot()
	in.isOr = q.getOr()
	q.addComponent(in)
	return q
}

func (q *query) WhereNotIn(columnName string, members ...interface{}) *query {
	return q.Not().WhereIn(columnName, members...)
}

func (q *query) OrWhereIn(columnName string, members ...interface{}) *query {
	return q.Or().WhereIn(columnName, members...)
}

func (q *query) OrWhereNotIn(columnName string, members ...interface{}) *query {
	return q.Or().Not().WhereIn(columnName, members...)
}

func (q *query) WhereNull(columnName string) *query {
	// cla := new(NullCondition)
	var cla nullCondition
	cla.columnName = columnName
	cla.componentName = "where"
	cla.isNot = q.getNot()
	cla.isOr = q.getOr()
	q.addComponent(cla)
	return q
}

func (q *query) WhereNotNull(columnName string) *query {
	return q.Not().WhereNull(columnName)
}

func (q *query) OrWhereNull(columnName string) *query {
	return q.Or().WhereNull(columnName)
}

func (q *query) OrWhereNotNull(columnName string) *query {
	return q.Or().Not().WhereNull(columnName)
}

func (q *query) WhereTrue(columnName string) *query {
	// cla := new(booleanCondition)
	var cla booleanCondition
	cla.columnName = columnName
	cla.value = true
	cla.componentName = "where"
	cla.isNot = q.getNot()
	cla.isOr = q.getOr()
	q.addComponent(cla)
	return q
}

func (q *query) OrWhereTrue(columnName string) *query {
	return q.Or().WhereTrue(columnName)
}

func (q *query) WhereFalse(columnName string) *query {
	// cla := new(booleanCondition)
	var cla booleanCondition
	cla.columnName = columnName
	cla.value = true
	cla.componentName = "where"
	cla.isNot = q.getNot()
	cla.isOr = q.getOr()
	q.addComponent(cla)
	return q
}

func (q *query) OrWhereFalse(columnName string) *query {
	return q.Or().WhereFalse(columnName)
}

func (q *query) OrderBy(columnName string) *query {
	// cla := new(orderByClause)
	var cla orderByClause
	cla.columnName = columnName
	cla.componentName = "order"
	q.addComponent(cla)
	return q
}

func (q *query) OrderByDesc(columnName string) *query {
	// cla := new(orderByClause)
	var cla orderByClause
	cla.columnName = columnName
	cla.desc = true
	cla.componentName = "order"
	q.addComponent(cla)
	return q
}

func (q *query) GroupBy(columnNames ...string) *query {
	// cla := new(groupByClause)
	var cla groupByClause
	cla.columnNames = columnNames
	cla.componentName = "group"
	q.replaceOrAdd(cla)
	return q
}

func (q *query) Having(columnName string, sign string, value interface{}) *query {
	var cla compareCondition
	cla.columnName = columnName
	cla.sign = sign
	cla.value = value
	cla.componentName = "having"
	cla.isNot = q.getNot()
	cla.isOr = q.getOr()
	q.addComponent(cla)
	return q
}

func (q *query) OrHaving(columnName string, sign string, value interface{}) *query {
	return q.Or().Having(columnName, sign, value)
}

func (q *query) HavingRaw(expression string) *query {
	var cla rawCodition
	cla.expression = expression
	cla.componentName = "having"
	cla.isNot = q.getNot()
	cla.isOr = q.getOr()
	q.addComponent(cla)
	return q
}

func (q *query) OrHavingRaw(expression string) *query {
	return q.Or().HavingRaw(expression)
}

func (q *query) Limit(rowCount int) *query {
	// cla := new(limitClause)
	var cla limitClause
	if rowCount < 0 {
		cla.rowCount = 0
	} else {
		cla.rowCount = int(rowCount)
	}
	cla.componentName = "limit"
	q.replaceOrAdd(cla)
	return q
}

func (q *query) Offset(rows int) *query {
	// cla := new(offsetClause)
	var cla offsetClause
	if rows < 0 {
		cla.offset = 0
	} else {
		cla.offset = int(rows)
	}
	cla.componentName = "offset"
	q.replaceOrAdd(cla)
	return q
}

func (q *query) WhereInQuery(columnName string, subQuery *query) *query {
	var cla inQueryCondition
	cla.columnName = columnName
	cla.subQuery = subQuery
	cla.componentName = "where"
	cla.isNot = q.getNot()
	cla.isOr = q.getOr()
	q.addComponent(cla)
	return q
}

func (q *query) OrWhereInQuery(columnName string, subQuery *query) *query {
	return q.Or().WhereInQuery(columnName, subQuery)
}

func (q *query) WhereNotInQuery(columnName string, subQuery *query) *query {
	return q.Not().WhereInQuery(columnName, subQuery)
}

func (q *query) OrWhereNotInQuery(columnName string, subQuery *query) *query {
	return q.Or().Not().WhereInQuery(columnName, subQuery)
}

func (q *query) WhereExists(subQuery *query) *query {
	var cla existsCondition
	cla.subQuery = subQuery
	cla.componentName = "where"
	cla.isNot = q.getNot()
	cla.isOr = q.getOr()
	q.addComponent(cla)
	return q
}

func (q *query) WhereNotExists(subQuery *query) *query {
	return q.Not().WhereExists(subQuery)
}

func (q *query) join(typ joinType, tableName, leftTable, sign, rightTable string) *query {
	var cla joinClause
	cla.joinTyp = typ
	cla.left = leftTable
	cla.right = rightTable
	cla.table = tableName
	cla.sign = sign
	cla.componentName = "join"
	q.addComponent(cla)
	return q
}

func (q *query) LeftJoin(tableName, leftTable, sign, rightTable string) *query {
	return q.join(leftJoin, tableName, leftTable, sign, rightTable)
}

func (q *query) RightJoin(tableName, leftTable, sign, rightTable string) *query {
	return q.join(rightJoin, tableName, leftTable, sign, rightTable)
}

func (q *query) Join(tableName, leftTable, sign, rightTable string) *query {
	return q.join(innerJoin, tableName, leftTable, sign, rightTable)
}

func (q *query) addComponent(cpt component) bool {
	q.components = append(q.components, cpt)
	return true
}

func (q *query) replaceComponent(cpt component) int {
	n := 0
	for i := 0; i < len(q.components); i++ {
		c := q.components[i]
		if c.getComponentName() == cpt.getComponentName() {
			q.components[i] = cpt
			n++
		}
	}
	return n
}

func (q *query) replaceOrAdd(cpt component) int {
	var n int
	if n := q.replaceComponent(cpt); n == 0 {
		q.addComponent(cpt)
		return 1
	}
	return n
}

func (q *query) getNot() bool {
	isNot := q.flagNot
	q.flagNot = false
	return isNot
}

func (q *query) getOr() bool {
	isOr := q.flagOr
	q.flagOr = false
	return isOr
}

func (q *query) getComponent(name string) (component, bool) {
	for i := range q.components {
		if q.components[i].getComponentName() == name {
			return q.components[i], true
		}
	}
	return nil, false
}

func (q *query) getComponents(componentName string) ([]component, int) {
	// cpts := make([]component, 0)
	var cpts []component
	for i := range q.components {
		if q.components[i].getComponentName() == componentName {
			cpts = append(cpts, q.components[i])
		}
	}
	return cpts, len(cpts)
}
