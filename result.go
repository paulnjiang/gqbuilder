package gqbuilder

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type sqlArguments struct {
	values       []interface{}
	symbolPrefix string
	pattern      bindPattern
}

func newSQLArguments(bp bindPattern, symbol string) *sqlArguments {
	p := new(sqlArguments)
	p.values = make([]interface{}, 0, 16)
	p.symbolPrefix = symbol
	p.pattern = bp
	return p
}

func (s *sqlArguments) Set(value interface{}) string {
	switch s.pattern {
	case PlaceHolder:
		s.values = append(s.values, value)
		return s.symbolPrefix
	case Ordinal:
		s.values = append(s.values, value)
		return s.symbolPrefix + strconv.Itoa(len(s.values))
	case Naming:
		p := len(s.values)
		if nam, ok := value.(sql.NamedArg); ok {
			s.values = append(s.values, nam)
			return s.symbolPrefix + nam.Name
		}
		nam := "param" + strconv.Itoa(p)
		s.values = append(s.values, sql.Named(nam, value))
		return s.symbolPrefix + nam
		
	default:
		s.values = append(s.values, value)
		return s.symbolPrefix
	}
}

func (s *sqlArguments) SetNameValue(name string, value interface{}) string {
	s.values = append(s.values, sql.Named(name, value))
	return s.symbolPrefix + name
}

func (s *sqlArguments) Clean() {
	s.values = make([]interface{}, 0, 16)
}

func (s *sqlArguments) Len() int {
	return len(s.values)
}

func (s *sqlArguments) GetByIndex(index int) interface{} {
	return s.values[index]
}

func (s *sqlArguments) GetByName(name string) (interface{}, bool) {
	if s.pattern != Naming {
		return nil, false
	}

	n := s.Len()
	for i := 0; i < n; i++ {
		argv := s.GetByIndex(i)
		if v, ok := argv.(sql.NamedArg); ok {
			if v.Name == name {
				return v.Value, true
			}
		}
		continue
	} 
	return nil, false
}
 
// SQLResult is a type that return from compiler's compile function
type SQLResult struct {
	args   *sqlArguments
	sql    string
	rawSQL string
}

func newSQLResult(bp bindPattern, symbol string) *SQLResult {
	ps := newSQLArguments(bp, symbol)
	res := new(SQLResult)
	res.args = ps
	return res
}

// ToString convert result to sql statement and replace bind variables placeholder
func (s *SQLResult) ToString() (string, error) {
	if s.sql != "" {
		return s.sql, nil
	}
	ssql := s.rawSQL
	n := s.args.Len()
	switch s.args.pattern {
	case PlaceHolder:
		for i := 0; i < n; i++ {
			refv := s.args.GetByIndex(i)
			if v, ok := refv.(string); ok {
				v = "'" + v + "'"
				ssql = strings.Replace(ssql, s.args.symbolPrefix, v, 1)
				continue
			}
			if v, ok := refv.(bool); ok {
				if v {
					ssql = strings.Replace(ssql, s.args.symbolPrefix, "TRUE", 1)
				} else {
					ssql = strings.Replace(ssql, s.args.symbolPrefix, "FALSE", 1)
				}
				continue
			}
			if v, ok := s.numberToString(refv); ok {
				ssql = strings.Replace(ssql, s.args.symbolPrefix, v, 1)
				continue
			}
			if v, ok := refv.(time.Time); ok {
				ssql = strings.Replace(ssql, s.args.symbolPrefix, v.String(), 1)
				continue
			}
			return "", fmt.Errorf("argument %v(%T) can not be coverted to string", refv, refv)
		}
	case Ordinal:
		for i := 0; i < n; i++ {
			refv := s.args.GetByIndex(i)
			old := s.args.symbolPrefix + strconv.Itoa(i+1)
			if p, ok := refv.(string); ok {
				p = "'" + p + "'"
				ssql = strings.Replace(ssql, old, p, 1)
				continue
			}
			if p, ok := refv.(bool); ok {
				if p {
					ssql = strings.Replace(ssql, old, "TRUE", 1)

				} else {
					ssql = strings.Replace(ssql, old, "FALSE", 1)
				}
				continue
			}
			if p, ok := s.numberToString(refv); ok {
				ssql = strings.Replace(ssql, old, p, 1)
				continue
			}
			if v, ok := refv.(time.Time); ok {
				ssql = strings.Replace(ssql, old, v.String(), 1)
				continue
			}	
			return "", fmt.Errorf("argument %v(%T) can not be coverted to string", refv, refv)
			
		}
	case Naming:
		for i := 0; i < n; i++ {
			refv := s.args.GetByIndex(i)
			if nv, ok := refv.(sql.NamedArg); ok {
				old := s.args.symbolPrefix + nv.Name
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
				if v, ok := s.numberToString(nv.Value); ok {
					ssql = strings.Replace(ssql, old, v, 1)
					continue
				}
				if v, ok := refv.(time.Time); ok {
					ssql = strings.Replace(ssql, old, v.String(), 1)
					continue
				}
				return "", fmt.Errorf("argument %#v(%T) can not be coverted to string", refv, refv)
			}
		}
	}
	s.sql = ssql
	return ssql, nil
}

// ToPrepared convert result to a sql statment with placeholders, and a bind variables list
func (s *SQLResult) ToPrepared() (string, []interface{}) {
	return s.rawSQL, s.args.values
}

func (s *SQLResult) numberToString(any interface{}) (string, bool) {
	switch v := any.(type) {
	case int:
		return strconv.Itoa(v), true
	case int32:
		return strconv.FormatInt(int64(v), 10), true
	case int8:
		return strconv.FormatInt(int64(v), 10), true
	case int16:
		return strconv.FormatInt(int64(v), 10), true
	case int64:
		return strconv.FormatInt(v, 10), true
	case uint8:
		return strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return strconv.FormatUint(uint64(v), 10), true
	case uint32:
		return strconv.FormatUint(uint64(v), 10), true
	case uint64:
		return strconv.FormatUint(v, 10), true
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), true
	case float64:
		return strconv.FormatFloat(float64(v), 'f', -1, 64), true
	default:
		return "", false
	}
}
