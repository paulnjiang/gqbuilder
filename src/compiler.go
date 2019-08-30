/*
	common tools
*/

package gqbuilder

import (
	"reflect"
	"strconv"
	"strings"
)

type compiler interface {
	compile(q *query) *SQLResult
}

func compilerFactory(engine databaseType) compiler {
	switch engine {
	case SQLite:
		return newSQLiteCompiler()
	case MySQL:
		return newMySQLCompiler()
	case PostgreSQL:
		return newPostgreSQLCompiler()
	default:
		return newCompiler()
	}
}

type baseCompiler struct {
	paramsPattern  varsPattern
	symbolPrefix   string
	leftIdentifier string
	righIdentifier string
	result         *SQLResult
}

func newCompiler() *baseCompiler {
	c := new(baseCompiler)
	c.paramsPattern = PlaceHolder
	c.symbolPrefix = ":"
	c.leftIdentifier = "\""
	c.righIdentifier = "\""
	c.result = nil
	return c
}

func (c *baseCompiler) compile(q *query) *SQLResult {
	c.result = newSQLResult(c.paramsPattern, c.symbolPrefix)
	switch q.method {
	case "select":
		c.compileSelect(q)
	default:
		return c.result
	}
	return c.result
}

func (c *baseCompiler) clone() compiler {
	cc := *c
	cc.result = newSQLResult(c.paramsPattern, c.symbolPrefix)
	return &cc
}

func (c *baseCompiler) wrapKeyWord(word string) string {
	return c.leftIdentifier + word + c.righIdentifier
}

func (c *baseCompiler) setArgument(v interface{}) string {
	return c.result.params.Set(v)
}

func (c *baseCompiler) compileSelect(q *query) *SQLResult {
	sql := make([]string, 0, 16)
	if q.isDistinct {
		sql = append(sql, kwSELECT, kwDISTINCT)
	} else {
		sql = append(sql, kwSELECT)
	}
	sql = c.append(sql, 
		c.compileColumns(q),
		c.compileFrom(q),
		c.compileJoins(q),
		c.compileWheres(q),
		c.compileGroupBy(q),
		c.compileHaving(q),
		c.compileOrderBy(q),
		c.compileLimitOffset(q))

	c.result.rawSQL = strings.Join(sql, " ")
	return c.result
}

func (c *baseCompiler) compileColumns(q *query) string {
	cpts, n := q.getComponents("column")
	if n == 0 {
		return c.wrapKeyWord(kwALL)
	}
	cols := make([]string, 0, n)
	for i := 0; i < n; i++ {
		if cla, ok := cpts[i].(columnClause); ok {
			if cla.alias == "" {
				cols = append(cols, c.wrapKeyWord(cla.name))
			} else {
				cols = append(cols, c.wrapKeyWord(cla.name)+kwAS+c.wrapKeyWord(cla.alias))
			}
			continue
		}
		if cla, ok := cpts[i].(rawColumnClause); ok {
			cols = append(cols, c.wrapKeyWord(cla.expression))
		}
	}
	return strings.Join(cols, ", ")
}

func (c *baseCompiler) compileFrom(q *query) string {
	if cpt, ok := q.getComponent("from"); ok {
		cla := cpt.(fromClause)
		return kwFROM + " " + c.wrapKeyWord(cla.tableName)
	}
	return ""
}

func (c *baseCompiler) compileJoins(q *query) string {
	cpts, n := q.getComponents("join")
	if n == 0 {
		return ""
	}
	joins := make([]string, 0, n)
	for _, cpt := range cpts {
		cla := cpt.(joinClause)
		joins = append(joins, c.compileJoin(cla))
	}
	return strings.Join(joins, " ")
}

func (c *baseCompiler) compileJoin(cla joinClause) string {
	var stmt []string
	switch cla.joinTyp {
	case leftJoin:
		stmt = append(stmt, kwLEFTJOIN)
	case rightJoin:
		stmt = append(stmt, kwRIGHTJOIN)
	case innerJoin:
		stmt = append(stmt, kwINNERJOIN)
	default:
		stmt = append(stmt, kwJOIN)
	}
	stmt = append(stmt, c.wrapKeyWord(cla.table), kwON, c.wrapKeyWord(cla.left), cla.sign, c.wrapKeyWord(cla.right))
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) compileWheres(q *query) string {
	cpts, n := q.getComponents("where")
	if n == 0 {
		return ""
	}

	stmt := make([]string, 0, n*2)
	stmt = append(stmt, kwWHERE)
	refv := reflect.ValueOf(c)
	if !refv.IsValid() {
		panic("bad compiler(zero value)")
	}
	for i := 0; i < n; i++ {
		cond := cpts[i]
		if i > 0 {
			isOr := reflect.ValueOf(cond).FieldByName("isOr")
			if !isOr.IsValid() {
				continue
			}
			if isOr.Bool() {
				stmt = append(stmt, kwOR)
			} else {
				stmt = append(stmt, kwAND)
			}
		}

		clauseName := reflect.TypeOf(cond).Name()
		title := strings.Title(strings.Replace(clauseName, "Condition", "", -1))
		methodName := "Compile" + title
		m := refv.MethodByName(methodName)
		if !m.IsValid() {
			continue
		}
		retVals := m.Call([]reflect.Value{reflect.ValueOf(cond)})
		stmt = append(stmt, retVals[0].String())
	}
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) compileGroupBy(q *query) string {
	if cpt, ok := q.getComponent("group"); ok {
		cla := cpt.(groupByClause)
		n := len(cla.columnNames)
		cols := make([]string, 0, n)
		for i := 0; i < n; i++ {
			cols = append(cols, c.wrapKeyWord(cla.columnNames[i]))
		}
		return kwGROUPBY + " " + strings.Join(cols, ", ")
	}
	return ""
}

func (c *baseCompiler) compileOrderBy(q *query) string {
	cpts, n := q.getComponents("order")
	if n == 0 {
		return ""
	}
	cols := make([]string, 0, n)
	for i := 0; i < n; i++ {
		cla := cpts[i].(orderByClause)
		if cla.desc {
			cols = append(cols, c.wrapKeyWord(cla.columnName)+" "+kwDESC)
		} else {
			cols = append(cols, c.wrapKeyWord(cla.columnName)+" "+kwASC)
		}
	}
	return kwORDERBY + kwSPACE + strings.Join(cols, ", ")
}

func (c *baseCompiler) compileLimitOffset(q *query) string {
	cpt, ok := q.getComponent("limit")
	if !ok {
		return ""
	}
	limit := cpt.(limitClause)
	if limit.rowCount <= 0 {
		return ""
	}

	stmt := make([]string, 0, 4)
	stmt = append(stmt, kwLIMIT, strconv.Itoa(limit.rowCount))

	if cpt, ok := q.getComponent("offset"); ok {
		ofs := cpt.(offsetClause)
		if ofs.offset > 0 {
			stmt = append(stmt, kwOFFSET, strconv.Itoa(ofs.offset))
		} else {
			stmt = append(stmt, kwOFFSET, "0")
		}
	}
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) CompileCompare(cpt component) string {
	cond := cpt.(compareCondition)
	ph := c.setArgument(cond.value)
	stmt := []string{c.wrapKeyWord(cond.columnName), cond.sign, ph}
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) CompileColumnCompare(cpt component) string {
	cond := cpt.(columnCompareCondition)
	stmt := []string{c.wrapKeyWord(cond.leftColumn), cond.sign, c.wrapKeyWord(cond.rightColumn)}
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) CompileBetween(cpt component) string {
	cond := cpt.(betweenCondition)
	f := c.setArgument(cond.from)
	t := c.setArgument(cond.to)
	if cond.isNot {
		stmt := []string{c.wrapKeyWord(cond.columnName), kwNOT, kwBETWEEN, f, kwAND, t}
		return strings.Join(stmt, " ")
	}
	stmt := []string{c.wrapKeyWord(cond.columnName), kwBETWEEN, f, kwAND, t}
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) CompileIn(cpt component) string {
	cond := cpt.(inCondition)
	var stmt []string
	if cond.isNot {
		stmt = []string{c.wrapKeyWord(cond.columnName), kwNOT, kwIN, "("}
	} else {
		stmt = []string{c.wrapKeyWord(cond.columnName), kwIN, "("}
	}
	var mbrs []string
	for mbr := range cond.members {
		ph := c.setArgument(mbr)
		mbrs = append(mbrs, ph)
	}
	stmt = append(stmt, strings.Join(mbrs, ", "), ")")
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) CompileNull(cpt component) string {
	cond := cpt.(nullCondition)
	if cond.isNot {
		stmt := []string{c.wrapKeyWord(cond.columnName), kwIS, kwNOT, kwNULL}
		return strings.Join(stmt, " ")
	}
	stmt := []string{c.wrapKeyWord(cond.columnName), kwIS, kwNULL}
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) CompileBoolean(cpt component) string {
	cond := cpt.(booleanCondition)
	if cond.isNot {
		stmt := []string{c.wrapKeyWord(cond.columnName), "=", kwFALSE}
		return strings.Join(stmt, " ")
	}
	stmt := []string{c.wrapKeyWord(cond.columnName), "=", kwTRUE}
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) CompileRaw(cpt component) string {
	cond := cpt.(rawCodition)
	return cond.expression
}

func (c *baseCompiler) CompileExists(cpt component) string {
	cond := cpt.(existsCondition)
	cc := c.clone()
	subq, _ := cc.compile(cond.subQuery).ToString()
	if cond.isNot {
		return kwNOT + " " + kwEXISTS + " (" + subq + ") "
	}
	return kwEXISTS + " (" + subq + ") "
}

func (c *baseCompiler) CompileInQuery(cpt component) string {
	stmt := make([]string, 0, 8)
	cond := cpt.(inQueryCondition)
	stmt = append(stmt, c.wrapKeyWord(cond.columnName))
	cc := c.clone()
	subq, _ := cc.compile(cond.subQuery).ToString()
	subq = "(" + subq + ")"
	if cond.isNot {
		stmt = append(stmt, kwNOT, kwIN, subq)
	} else {
		stmt = append(stmt, kwIN, subq)
	}
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) compileHaving(q *query) string {
	cpts, n := q.getComponents("having")
	if n == 0 {
		return ""
	}

	stmt := make([]string, 0, n*2)
	stmt = append(stmt, kwHAVING)
	for i := 0; i < n; i++ {
		switch cond := cpts[i].(type) {
		case compareCondition:
			if i > 0 {
				if cond.isOr {
					stmt = append(stmt, kwOR)
				} else {
					stmt = append(stmt, kwAND)
				}
			}
			stmt = append(stmt, c.CompileCompare(cpts[i]))
		case rawCodition:
			if i > 0 {
				if cond.isOr {
					stmt = append(stmt, kwOR)
				} else {
					stmt = append(stmt, kwAND)
				}
			}
			stmt = append(stmt, c.CompileRaw(cpts[i]))
		default:
			continue
		}
	}
	return strings.Join(stmt, " ")
}

func (c *baseCompiler) append(sli []string, str string, otherStrs ...string) []string {
	if str != "" {
		sli = append(sli, str)
	}
	for _, s := range otherStrs {
		if s == "" {
			continue
		}
		sli = append(sli, s)
	}
	return sli
} 