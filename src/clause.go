package gqbuilder

/*
	A simple SQL builder
*/

type component interface {
	getComponentName() string
}

// type conditionComponent interface {
// 	component
// 	isAndOperator() bool
// }

type baseClause struct {
	componentName string
}

func (b baseClause) getComponentName() string {
	return b.componentName
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
	subQuery   *query
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
	subQuery *query
	conditionClause
}

type rawCodition struct {
	expression string
	conditionClause
}
