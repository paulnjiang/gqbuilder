/*
	common tools
*/

package gqbuilder

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// CompileError records an error and function that caused it
type CompileError struct {
	Function string
	Err      error
}

func (c *CompileError) Error() string {
	return c.Function + ": " + c.Err.Error()
}

func (c *CompileError) Unwrap() string {
	return c.Err.Error()
}

type compiler interface {
	compile(q *Query) (*SQLResult, error)
	clone() compiler
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
		return newBaseCompiler()
	}
}

type baseCompiler struct {
	paramsPattern  bindPattern
	symbolPrefix   string
	leftIdentifier string
	righIdentifier string
	result         *SQLResult
}

func newBaseCompiler() *baseCompiler {
	c := new(baseCompiler)
	c.paramsPattern = PlaceHolder
	c.symbolPrefix = "?"
	c.leftIdentifier = "\""
	c.righIdentifier = "\""
	c.result = nil
	return c
}

func (c *baseCompiler) compile(q *Query) (*SQLResult, error) {
	// first execute
	if c.result == nil {
		c.result = newSQLResult(c.paramsPattern, c.symbolPrefix)
	}
	switch q.method {
	case selectMethod:
		return c.result, c.CompileSelect(q)
	case insertMethod:
		return c.result, c.CompileInsert(q)
	case updateMethod:
		return c.result, c.CompileUpdate(q)
	case deleteMethod:
		return c.result, c.CompileDelete(q)
	default:
		return c.result, &CompileError{"compile: ", errors.New("query method type error")}
	}
}

func (c *baseCompiler) clone() compiler {
	cc := *c
	return &cc
}

func (c *baseCompiler) wrapWord(word string) string {
	return c.leftIdentifier + word + c.righIdentifier
}

func (c *baseCompiler) append(slc []string, str string, otherStrs ...string) []string {
	if str != "" {
		slc = append(slc, str)
	}
	for _, s := range otherStrs {
		if s == "" {
			continue
		}
		slc = append(slc, s)
	}
	return slc
}

func (c *baseCompiler) setArgument(v interface{}) string {
	return c.result.args.Set(v)
}

func (c *baseCompiler) CompileHaving(q *Query) (string, error) {
	cpns, n := q.getElements("having")
	if n == 0 {
		return "", nil
	}
	stmt := make([]string, 0, n*2)
	stmt = append(stmt, kwHAVING)
	for i := 0; i < n; i++ {
		switch cond := cpns[i].(type) {
		case compareCondition:
			if i > 0 {
				if cond.isOr {
					stmt = append(stmt, kwOR)
				} else {
					stmt = append(stmt, kwAND)
				}
			}
			s, err := c.CompileCompare(cpns[i])
			if err != nil {
				return "", &CompileError{"compileHaving", err}
			}
			stmt = append(stmt, s)
		case rawCodition:
			if i > 0 {
				if cond.isOr {
					stmt = append(stmt, kwOR)
				} else {
					stmt = append(stmt, kwAND)
				}
			}
			raw, err := c.CompileRaw(cpns[i])
			if err != nil {
				return "", err
			}
			stmt = append(stmt, raw)
		default:
			continue
		}
	}
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileSelect(q *Query) error {
	stmt := make([]string, 0, 16)
	if q.isDistinct {
		stmt = append(stmt, kwSELECT, kwDISTINCT)
	} else {
		stmt = append(stmt, kwSELECT)
	}
	rst, err := c.CompileColumns(q)
	if err != nil {
		return &CompileError{"CompileSelect: ", err}
	}
	if rst != "" {
		stmt = append(stmt, rst)
	}
	rst, err = c.CompileFrom(q)
	if err != nil {
		return &CompileError{"CompileSelect: ", err}
	}
	if rst != "" {
		stmt = append(stmt, rst)
	}
	rst, err = c.CompileJoins(q)
	if err != nil {
		return &CompileError{"CompileSelect: ", err}
	}
	if rst != "" {
		stmt = append(stmt, rst)
	}
	rst, err = c.CompileWheres(q)
	if err != nil {
		return &CompileError{"CompileSelect: ", err}
	}
	if rst != "" {
		stmt = append(stmt, rst)
	}
	rst, err = c.CompileGroupBy(q)
	if err != nil {
		return &CompileError{"CompileSelect: ", err}
	}
	if rst != "" {
		stmt = append(stmt, rst)
	}
	rst, err = c.CompileHaving(q)
	if err != nil {
		return &CompileError{"CompileSelect: ", err}
	}
	if rst != "" {
		stmt = append(stmt, rst)
	}
	rst, err = c.CompileOrderBy(q)
	if err != nil {
		return &CompileError{"CompileSelect: ", err}
	}
	if rst != "" {
		stmt = append(stmt, rst)
	}
	rst, err = c.CompileLimit(q)
	if err != nil {
		return &CompileError{"CompileSelect: ", err}
	}
	if rst != "" {
		stmt = append(stmt, rst)
	}
	rst, err = c.CompileOffset(q)
	if err != nil {
		return &CompileError{"CompileSelect: ", err}
	}
	if rst != "" {
		stmt = append(stmt, rst)
	}
	c.result.rawSQL = strings.Join(stmt, kwSPACE)
	return nil
}

func (c *baseCompiler) CompileColumns(q *Query) (string, error) {
	clms := make([]string, 0)
	types := []string{"column", "RawColumn"}
	for _, t := range types {
		cpns, n := q.getElements(t)
		if n != 0 {
			for i := 0; i < n; i++ {
				switch cpns[i].(type) {
				case columnClause:
					cls := cpns[i].(columnClause)
					if cls.alias == "" {
						clms = append(clms, c.wrapWord(cls.name))
					} else {
						as := c.wrapWord(cls.name) + kwAS + c.wrapWord(cls.alias)
						clms = append(clms, as)
					}
				case rawColumnClause:
					cls := cpns[i].(rawColumnClause)
					clms = append(clms, cls.expression)
				default:
					continue
				}
			}
		}
	}
	return strings.Join(clms, kwCOMMA), nil
}

func (c *baseCompiler) CompileFrom(q *Query) (string, error) {
	var stmt []string
	elms, n := q.getElements("from")
	if n == 0 {
		return "", errors.New("no table specified")
	}
	for _, elm := range elms {
		cls, ok := elm.(fromClause)
		if !ok {
			return "", &CompileError{"compileForm", errors.New("assert error")}
		}
		
		if cls.alias != "" {
			stmt = append(stmt, c.wrapWord(cls.tableName) + kwAS + c.wrapWord(cls.alias))
		} else {
			stmt = append(stmt, c.wrapWord(cls.tableName))
		}
	}
	return kwFROM + kwSPACE + strings.Join(stmt, kwCOMMA), nil
}

func (c *baseCompiler) CompileJoins(q *Query) (string, error) {
	cpns, n := q.getElements("join")
	if n == 0 {
		return "", nil
	}
	joins := make([]string, 0, n)
	for _, elm := range cpns {
		cls, ok := elm.(joinClause)
		if !ok {
			return "", &CompileError{"compileForm", errors.New("assert error")}
		}
		joins = append(joins, c.CompileJoin(cls))
	}
	return strings.Join(joins, kwSPACE), nil
}

func (c *baseCompiler) CompileJoin(cla joinClause) string {
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
	stmt = append(stmt, c.wrapWord(cla.table), kwON, c.wrapWord(cla.left), cla.sign, c.wrapWord(cla.right))
	return strings.Join(stmt, kwSPACE)
}

func (c *baseCompiler) CompileWheres(q *Query) (string, error) {
	cpns, n := q.getElements("where")
	if n == 0 {
		return "", nil
	}

	stmt := make([]string, 0, n*2)
	stmt = append(stmt, kwWHERE)
	refv := reflect.ValueOf(c)
	if !refv.IsValid() {
		panic("bad compiler(zero value)")
	}
	for i := 0; i < n; i++ {
		cpnt := cpns[i]
		if i > 0 {
			isOr := reflect.ValueOf(cpnt).FieldByName("isOr")
			if !isOr.IsValid() {
				continue
			}
			if isOr.Bool() {
				stmt = append(stmt, kwOR)
			} else {
				stmt = append(stmt, kwAND)
			}
		}

		clauseName := reflect.TypeOf(cpnt).Name()
		title := strings.Title(strings.Replace(clauseName, "Condition", "", -1))
		methodName := "Compile" + title
		m := refv.MethodByName(methodName)
		if !m.IsValid() {
			continue
		}
		retVals := m.Call([]reflect.Value{reflect.ValueOf(cpnt)})
		err := retVals[1].Interface()

		if err != nil {
			switch err.(type) {
			case error:
				e := err.(error)
				return strings.Join(stmt, kwSPACE), e
			default:
				return strings.Join(stmt, kwSPACE), &CompileError{"compileWheres: ", errors.New("unknow error")}
			}
		}
		stmt = append(stmt, retVals[0].String())
	}
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileGroupBy(q *Query) (string, error) {
	if elm, ok := q.getElement("group"); ok {
		cls, ok := elm.(groupByClause)
		if !ok {
			return "", &CompileError{"compileGroupBy", errors.New("assert error")}
		}
		n := len(cls.columnNames)
		cols := make([]string, 0, n)
		for i := 0; i < n; i++ {
			cols = append(cols, c.wrapWord(cls.columnNames[i]))
		}
		return kwGROUPBY + kwSPACE + strings.Join(cols, kwCOMMA), nil
	}
	return "", nil
}

func (c *baseCompiler) CompileOrderBy(q *Query) (string, error) {
	cpns, n := q.getElements("order")
	if n == 0 {
		return "", nil
	}
	cols := make([]string, 0, n)
	for i := 0; i < n; i++ {
		cls, ok := cpns[i].(orderByClause)
		if !ok {
			return "", &CompileError{"compileOrderBy", errors.New("assert error")}
		}
		if cls.desc {
			cols = append(cols, c.wrapWord(cls.columnName) + kwSPACE + kwDESC)
		} else {
			cols = append(cols, c.wrapWord(cls.columnName) + kwSPACE + kwASC)
		}
	}
	return kwORDERBY + kwSPACE + strings.Join(cols, kwCOMMA), nil
}

func (c *baseCompiler) CompileLimit(q *Query) (string, error) {
	elm, ok := q.getElement("limit")
	if !ok {
		return "", nil
	}
	limit, ok := elm.(limitClause)
	if !ok {
		return "", &CompileError{"compileLimit", errors.New("assert error")}
	}
	if limit.rowCount <= 0 {
		return "", errors.New("limit must great than 0")
	}

	stmt := make([]string, 0, 2)
	stmt = append(stmt, kwLIMIT, strconv.Itoa(limit.rowCount))
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileOffset(q *Query) (string, error) {
	elm, ok := q.getElement("offset")
	if !ok {
		return "", nil
	}
	ofs, ok := elm.(offsetClause)
	if !ok {
		return "", &CompileError{"compileOffset", errors.New("assert error")}
	}
	stmt := make([]string, 0, 2)
	if ofs.offset > 0 {
		stmt = append(stmt, kwOFFSET, strconv.Itoa(ofs.offset))
	} else {
		// stmt = append(stmt, kwOFFSET, "0")
		return "", errors.New("offset must great than 0")
	}
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileCompare(elm element) (string, error) {
	cond, ok := elm.(compareCondition)
	if !ok {
		return "", &CompileError{"CompileCompare", errors.New("assert error")}
	}
	ph := c.setArgument(cond.value)

	// TODO: check sign
	stmt := []string{c.wrapWord(cond.columnName), cond.sign, ph}
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileLike(elm element) (string, error) {
	cond, ok := elm.(likeCondition)
	if !ok {
		return "", &CompileError{"CompileLike", errors.New("assert error")}
	}
	ph := c.setArgument(cond.like)
	if cond.isNot {
		stmt := []string{c.wrapWord(cond.columnName), kwNOT, kwLIKE, ph}
		return strings.Join(stmt, kwSPACE), nil
	}
	stmt := []string{c.wrapWord(cond.columnName), kwLIKE, ph}
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileColumnCompare(elm element) (string, error) {
	cond, ok := elm.(columnCompareCondition)
	if !ok {
		return "", &CompileError{"CompileColumnCompare", errors.New("assert error")}
	}
	stmt := []string{c.wrapWord(cond.leftColumn), cond.sign, c.wrapWord(cond.rightColumn)}
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileBetween(elm element) (string, error) {
	cond, ok := elm.(betweenCondition)
	if !ok {
		return "", &CompileError{"CompileBetween", errors.New("assert error")}
	}
	f := c.setArgument(cond.from)
	t := c.setArgument(cond.to)
	if cond.isNot {
		stmt := []string{c.wrapWord(cond.columnName), kwNOT, kwBETWEEN, f, kwAND, t}
		return strings.Join(stmt, kwSPACE), nil
	}
	stmt := []string{c.wrapWord(cond.columnName), kwBETWEEN, f, kwAND, t}
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileIn(elm element) (string, error) {
	cond, ok := elm.(inCondition)
	if !ok {
		return "", &CompileError{"CompileIn", errors.New("assert error")}
	}
	var stmt []string
	if cond.isNot {
		stmt = []string{c.wrapWord(cond.columnName), kwNOT, kwIN, "("}
	} else {
		stmt = []string{c.wrapWord(cond.columnName), kwIN, "("}
	}
	var smbr []string
	for _, mbr := range cond.members {
		ph := c.setArgument(mbr)
		smbr = append(smbr, ph)
	}
	stmt = append(stmt, strings.Join(smbr, kwCOMMA), ")")
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileNull(elm element) (string, error) {
	cond, ok := elm.(nullCondition)
	if !ok {
		return "", &CompileError{"CompileNull", errors.New("assert error")}
	}
	if cond.isNot {
		stmt := []string{c.wrapWord(cond.columnName), kwIS, kwNOT, kwNULL}
		return strings.Join(stmt, kwSPACE), nil
	}
	stmt := []string{c.wrapWord(cond.columnName), kwIS, kwNULL}
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileBoolean(elm element) (string, error) {
	cond, ok := elm.(booleanCondition)
	if !ok {
		return "", &CompileError{"CompileBoolean", errors.New("assert error")}
	}
	if cond.isNot {
		stmt := []string{c.wrapWord(cond.columnName), "=", kwFALSE}
		return strings.Join(stmt, kwSPACE), nil
	}
	stmt := []string{c.wrapWord(cond.columnName), "=", kwTRUE}
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileRaw(elm element) (string, error) {
	cond, ok := elm.(rawCodition)
	if !ok {
		return "", &CompileError{"CompileRaw", errors.New("assert error")}
	}
	return cond.expression, nil
}

func (c *baseCompiler) CompileExists(elm element) (string, error) {
	cond, ok := elm.(existsCondition)
	if !ok {
		return "", &CompileError{"CompileExists", errors.New("assert error")}
	}
	rst, err := c.compile(cond.subQuery)
	if err != nil {
		return "", err
	}
	subq := rst.rawSQL
	if cond.isNot {
		return kwNOT + kwSPACE + kwEXISTS + " (" + subq + ") ", nil
	}
	return kwEXISTS + " (" + subq + ") ", nil
}

func (c *baseCompiler) CompileInQuery(elm element) (string, error) {
	stmt := make([]string, 0, 8)
	cond, ok := elm.(inQueryCondition)
	if !ok {
		return "", &CompileError{"CompileInQuery", errors.New("assert error")}
	}
	stmt = append(stmt, c.wrapWord(cond.columnName))
	rst, err := c.compile(cond.subQuery)
	if err != nil {
		return "", err
	}
	subq := "(" + rst.rawSQL + ")"
	if cond.isNot {
		stmt = append(stmt, kwNOT, kwIN, subq)
	} else {
		stmt = append(stmt, kwIN, subq)
	}
	return strings.Join(stmt, kwSPACE), nil
}

func (c *baseCompiler) CompileInsert(q *Query) error {
	var elm element
	stmt := []string{kwINSERT}
	elm, has := q.getElement("from")
	if !has {
		return &CompileError{"compileInsert", errors.New("no table specified")}
	}
	tableName := c.wrapWord(elm.(fromClause).tableName)
	stmt = append(stmt, tableName)

	elm, has = q.getElement("insert")
	if !has {
		return &CompileError{"compileInsert", errors.New("insert clause not exists")}
	}
	ic := elm.(insertClause)

	// from sub query
	if ic.subQuery != nil {
		str, e := ic.subQuery.ToString()
		if e != nil {
			return &CompileError{"compileInsert", e}
		} 
		stmt = append(stmt, str)
		
		c.result.rawSQL = strings.Join(stmt, kwSPACE)
		return nil
	}

	// name values OR only value
	if len(ic.columns) != 0 {
		cols := "(" + c.leftIdentifier
		cols = cols + strings.Join(ic.columns, c.leftIdentifier + kwCOMMA + c.righIdentifier)
		cols = cols + c.righIdentifier + ")"
		stmt = append(stmt, cols, kwVALUES)
	} else {
		stmt = append(stmt, kwVALUES)
	}

	// replace value to placeholder
	pls := make([]string, 0, len(ic.values))
	for _, v := range ic.values {
		pls = append(pls, c.setArgument(v))
	}
	stmt = append(stmt, "("+strings.Join(pls, kwCOMMA)+")")
	c.result.rawSQL = strings.Join(stmt, kwSPACE)
	return nil
}

func (c *baseCompiler) CompileUpdate(q *Query) error {
	var elm element
	stmt := []string{kwUPDATE}
	elm, has := q.getElement("from")
	if !has {
		return &CompileError{"compileUpdate", errors.New("no table specified")}
	}
	tableName := c.wrapWord(elm.(fromClause).tableName)
	stmt = append(stmt, tableName, kwSET)

	elm, has = q.getElement("update")
	if !has {
		return &CompileError{"compileUpdate", errors.New("update clause not exists")}
	}
	cls := elm.(updateClause)
	pairs := make([]string, 0)
	for clm, v := range cls.item {
		s := c.wrapWord(clm) + "=" + c.setArgument(v)
		pairs = append(pairs, s)
	}
	stmt = append(stmt, strings.Join(pairs, kwCOMMA))

	// compile condition clause
	whe, err := c.CompileWheres(q)
	if err != nil {
		return err
	}
	stmt = append(stmt, whe)
	c.result.rawSQL = strings.Join(stmt, kwSPACE)
	return nil
}

func (c *baseCompiler) CompileDelete(q *Query) error {
	var elm element
	stmt := []string{kwDELETE}
	elm, has := q.getElement("from")
	if !has {
		return &CompileError{"compileUpdate", errors.New("no table specified")}
	}
	tableName := c.wrapWord(elm.(fromClause).tableName)
	stmt = append(stmt, tableName)
	whe, err := c.CompileWheres(q)
	if err != nil {
		return err
	}
	stmt = append(stmt, whe)
	c.result.rawSQL = strings.Join(stmt, kwSPACE)
	return nil
}
