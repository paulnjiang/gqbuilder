package gqbuilder

/*
	A simple SQL builder
*/

type element interface {
	getElementName() string
}

type baseClause struct {
	elementName string
}

func (b baseClause) getElementName() string {
	return b.elementName
}

type conditionClause struct {
	isNot bool
	isOr  bool
	baseClause
}

func (c conditionClause) isAndOperator() bool {
	return !c.isOr
}

type columnClause struct {
	name  string
	alias string
	baseClause
}

type rawColumnClause struct {
	expression string
	baseClause
}

type fromClause struct {
	tableName string
	alias     string
	baseClause
}

type joinClause struct {
	table   string
	joinTyp joinType
	left    string
	right   string
	sign    string
	baseClause
}

type limitClause struct {
	rowCount int
	baseClause
}

type offsetClause struct {
	offset int
	baseClause
}

type orderByClause struct {
	columnName string
	desc       bool
	baseClause
}

type groupByClause struct {
	columnNames []string
	baseClause
}

type compareCondition struct {
	columnName string
	sign       string
	value      interface{}
	conditionClause
}

type columnCompareCondition struct {
	leftColumn  string
	rightColumn string
	sign        string
	conditionClause
}

type likeCondition struct {
	columnName string
	like       string
	conditionClause
}

type betweenCondition struct {
	columnName string
	from       interface{}
	to         interface{}
	conditionClause
}

type inCondition struct {
	columnName string
	members    []interface{}
	conditionClause
}

type inQueryCondition struct {
	columnName string
	subQuery   *Query
	conditionClause
}

type nullCondition struct {
	columnName string
	conditionClause
}

type booleanCondition struct {
	columnName string
	value      bool
	conditionClause
}

type existsCondition struct {
	subQuery *Query
	conditionClause
}

type rawCodition struct {
	expression string
	conditionClause
}

type insertClause struct {
	columns  []string
	values   []interface{}
	subQuery *Query
	baseClause
}

type updateClause struct {
	item map[string]interface{}
	baseClause
}
