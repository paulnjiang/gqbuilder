package gqbuilder

import (
	"testing"
	"fmt"
)

var i8 int8 = 12
var i16 int16 = 1000
var i32 int32 = -123
var i64 int64 = 100000000000000000
var ui8 uint8 = 13
var ui16 uint16 = 1000
var ui32 uint32 = 1000000000
var ui64 uint64 = 10000000000000000000
var f32 float32 = 3.1415
var f64 float64 = -3.14159267
var str = "asldkjflaksjdfaklj ajlksdjfa (0 asdfj a1223;;;:"
var tim = "1970/1/1 00:00:00"
var yes = true
var no = false

func TestPlaceholder(t *testing.T) {
	res := newSQLResult(PlaceHolder, "?")
	res.params.Set(i8)
	res.params.Set(i16)
	res.params.Set(i32)
	res.params.Set(i64)

	res.params.Set(ui8)
	res.params.Set(ui16)
	res.params.Set(ui32)
	res.params.Set(ui64)

	res.params.Set(f32)
	res.params.Set(f64)

	res.params.Set(str)
	res.params.Set(tim)

	res.params.Set(yes)
	res.params.Set(no)

	res.rawSQL = "1: ? 2: ? 3: ? 4: ? 5: ? 6: ? 7: ? 8: ? 9: ? 10: ? 11: ? 12: ? 13: ? 14: ?"
	if s, e := res.ToString(); e != nil {
		println(e)
		t.Fail()
	} else {
		println(s)
	}
}

func TestOrdinal(t *testing.T) {
	res := newSQLResult(Ordinal, "$")
	res.params.Set(i8)
	res.params.Set(i16)
	res.params.Set(i32)
	res.params.Set(i64)

	res.params.Set(ui8)
	res.params.Set(ui16)
	res.params.Set(ui32)
	res.params.Set(ui64)

	res.params.Set(f32)
	res.params.Set(f64)

	res.params.Set(str)
	res.params.Set(tim)

	res.params.Set(yes)
	res.params.Set(no)

	res.rawSQL = "1: $1 2: $2 3: $3 4: $4 5: $5 6: $6 7: $7 8: $8 9: $9 10: $10 11: $11 12: $12 13: $14 14: $13"
	if s, e := res.ToString(); e != nil {
		println(e)
		t.Fail()
	} else {
		println(s)
	}
}

type st struct {}
func TestNaming(t *testing.T) {
	r := st{}
	res := newSQLResult(Naming, ":")
	res.params.Set(i8)
	res.params.Set(i16)
	res.params.Set(i32)
	// res.params.Set(i64)
	res.params.Set(r)

	res.params.Set(ui8)
	res.params.Set(ui16)
	res.params.Set(ui32)
	res.params.Set(ui64)

	res.params.Set(f32)
	res.params.Set(f64)

	res.params.Set(str)
	res.params.Set(tim)

	res.params.Set(yes)
	res.params.Set(no)

	res.rawSQL = "1: :param0 2: :param1 3: :param2 4: :param3 5: :param4 6: :param5 7: :param6 8: :param7 9: :param8 10: :param9 11: :param10 12: :param11 13: :param13 14: :param12"
	if s, e := res.ToString(); e != nil {
		fmt.Printf("error: %v\n", e)
		t.Fail()
	} else {
		println(s)
	}
}