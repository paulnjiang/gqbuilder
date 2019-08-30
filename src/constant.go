package gqbuilder

/*
	const values
*/

type varsPattern int
type databaseType int
type joinType int

// Type of sql bind parameters
const (
	PlaceHolder varsPattern = iota // use placeholder e.g. '?'
	Naming                         // use prefix and parameter name e.g. '@parameterName'
	Ordinal                        // use prefix and positional e.g. '$1 $2'
)

// Type of database engine
const (
	SQLite databaseType = iota
	MySQL
	PostgreSQL
	// SQLServer
	// Oracle
)

const (
	leftJoin joinType = iota
	rightJoin
	innerJoin
	fullJoin
)

// Sql keywords
const (
	kwSELECT    string = "SELECT"
	kwUPDATE    string = "UPDATE"
	kwDELETE    string = "DELETE FROM"
	kwINSERT    string = "INSERT INTO"
	kwFROM      string = "FROM"
	kwJOIN      string = "JOIN"
	kwWHERE     string = "WHERE"
	kwSET       string = "SET"
	kwNOT       string = "NOT"
	kwLIKE      string = "LIKE"
	kwBETWEEN   string = "BETWEEN"
	kwIN        string = "IN"
	kwON        string = "ON"
	kwLIMIT     string = "LIMIT"
	kwOFFSET    string = "OFFSET"
	kwFETCH     string = "FETCH NEXT"
	kwORDERBY   string = "ORDER BY"
	kwGROUPBY   string = "GROUP BY"
	kwROWS      string = "ROWS"
	kwONLY      string = "ONLY"
	kwDESC      string = "DESC"
	kwASC       string = "ASC"
	kwDISTINCT  string = "DISTINCT"
	kwVALUES    string = "VALUES"
	kwIS        string = "IS"
	kwNULL      string = "NULL"
	kwAND       string = "AND"
	kwOR        string = "OR"
	kwAS        string = "AS"
	kwSPACE     string = " "
	kwALL       string = "*"
	kwCOMMA     string = ","
	kwFALSE     string = "False"
	kwTRUE      string = "True"
	kwHAVING    string = "HAVING"
	kwLEFTJOIN  string = "LEFT JOIN"
	kwRIGHTJOIN string = "RIGHT JOIN"
	kwINNERJOIN string = "INNER JOIN"
	kwEXISTS    string = "EXISTS"
)
