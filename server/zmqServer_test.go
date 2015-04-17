package magLogParserServer

import (

	"testing"
	
	)

func BenchmarkParseData1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDate1("2014/11/30 18:10:15")
	}
}

func BenchmarkParseData2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDate2("2014/11/30 18:10:15")
	}
}

func BenchmarkParseData3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDate3([]byte("2014/11/30 18:10:15"))
	}
}

func BenchmarkParseData4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDate4("2014/11/30 18:10:15")
	}
}
