package gqbuilder

import (
	"database/sql"
	"strconv"
	"errors"
	"strings"
)

type sqlParameters struct {
	values       []interface{}
	symbolPrefix string
	pattern      varsPattern
}

func newSQLParameters(bp varsPattern, symbol string) *sqlParameters {
	p := new(sqlParameters)
	p.values = make([]interface{}, 0, 16)
	p.symbolPrefix = symbol
	p.pattern = bp
	return p
}

func (s *sqlParameters) Set(value interface{}) string {
	switch s.pattern {
	case PlaceHolder:
		s.values = append(s.values, value)
		return s.symbolPrefix
	case Ordinal:
		s.values = append(s.values, value)
		return s.symbolPrefix + strconv.Itoa(len(s.values))
	case Naming:
		p := len(s.values)
		nam := "param" + strconv.Itoa(p)
		s.values = append(s.values, sql.Named(nam, value))
		return s.symbolPrefix + nam
	default:
		s.values = append(s.values, value)
		return s.symbolPrefix
	}
}

func (s *sqlParameters) SetNV(name string, value interface{}) string {
	s.values = append(s.values, sql.Named(name, value))
	return s.symbolPrefix + name
}

func (s *sqlParameters) Clean() {
	s.values = make([]interface{}, 0, 16)
}

// SQLResult is a type return from compiler's compile function
type SQLResult struct {
	params *sqlParameters
	sql    string
	rawSQL string
}

func newSQLResult(bp varsPattern, symbol string) *SQLResult {
	ps := newSQLParameters(bp, symbol)
	res := new(SQLResult)
	res.params = ps
	return res
}

// ToString convert result to sql statement and replace bind variables placeholder
func (s *SQLResult) ToString() (string, error) {
	if s.sql != "" {
		return s.sql, nil
	}
	ssql := s.rawSQL
	switch s.params.pattern {
	case PlaceHolder:
		for _, p := range s.params.values {
			if param, ok := p.(string); ok {
				param = "'" + param + "'"
				ssql = strings.Replace(ssql, s.params.symbolPrefix, param, 1)
				continue
			}
			if param, ok := p.(bool); ok {
				if param {
				ssql = strings.Replace(ssql, s.params.symbolPrefix, "TRUE", 1)
				
				} else {
					ssql = strings.Replace(ssql, s.params.symbolPrefix, "FALSE", 1)
				}
				continue
			}
			if param, e := s.numberToString(p); e == nil {
				ssql = strings.Replace(ssql, s.params.symbolPrefix, param, 1)
				continue
			} else {
				return "", e
			}
		}
	case Ordinal:
		params := s.params.values
		n := len(params)
		for i:=0; i<n; i++ {
			old := s.params.symbolPrefix + strconv.Itoa(i+1)
			if param, ok := params[i].(string); ok {
				param = "'" + param + "'"
				ssql = strings.Replace(ssql, old, param, 1)
				continue
			}
			if param, ok := params[i].(bool); ok {
				if param {
				ssql = strings.Replace(ssql, old, "TRUE", 1)
				
				} else {
					ssql = strings.Replace(ssql, old, "FALSE", 1)
				}
				continue
			}
			if param, e := s.numberToString(params[i]); e == nil {
				ssql = strings.Replace(ssql, old, param, 1)
				continue
			} else {
				return "", e
			}
		}
	case Naming:
		for _, param := range s.params.values {
			if nv, ok := param.(sql.NamedArg); ok {
				old := s.params.symbolPrefix + nv.Name
				if v, ok := nv.Value.(string); ok {
					v = "'" + v + "'"
					ssql = strings.Replace(ssql, old, v, 1)
					continue
				}
				if v, ok := nv.Value.(bool); ok {
					if v {
					ssql = strings.Replace(ssql, old, "TRUE", 1)
					
					} else {
						ssql = strings.Replace(ssql, old, "FALSE", 1)
					}
					continue
				}
				if v, e := s.numberToString(nv.Value); e == nil {
					ssql = strings.Replace(ssql, old, v, 1)
					continue
				} else {
					return "", e
				}
			}
		}
	}
	s.sql = ssql
	return ssql, nil
}

// ToPrepared convert result to a sql statment with placeholder, and a bind variables list
func (s *SQLResult) ToPrepared() (string, []interface{}) {
	return s.rawSQL, s.params.values
}

func (s *SQLResult) numberToString(any interface{}) (string, error) {
	switch v := any.(type) {
	case int:
		return strconv.Itoa(v), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(float64(v), 'f', -1, 64), nil
	default:
		return "", errors.New("unknown type")
	}
}