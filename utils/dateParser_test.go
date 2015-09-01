package utils

import (
	"testing"
)

func BenchmarkParseData1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDate1("2014/11/30 18:10:15.000000")
	}
}

func BenchmarkParseData2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDate2("2014/11/30 18:10:15.000000")
	}
}

func BenchmarkParseData3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDate3([]byte("2014/11/30 18:10:15.000000"))
	}
}

func BenchmarkParseData4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDate4("2014/11/30 18:10:15.000000")
	}
}
func TestParse4(t *testing.T) {
	s := "2015/08/04 15:45:04.975123"

	ti, _ := ParseDate4(s)

	if ti.Format(DateFormat) != s {
		t.Log(ti.Format(DateFormat))
		t.Fail()
	}
}
