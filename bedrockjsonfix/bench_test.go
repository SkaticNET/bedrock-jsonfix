package bedrockjsonfix

import (
	"strings"
	"testing"
)

var benchmarkResult Result

func BenchmarkFixBytesPreserveValid(b *testing.B) {
	input := []byte(`{"name":"stone","value":1}`)
	opt := DefaultOptions()
	b.SetBytes(int64(len(input)))

	for i := 0; i < b.N; i++ {
		res, err := FixBytes(input, opt)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkResult = res
	}
}

func BenchmarkFixBytesLargeValidNoPreserve(b *testing.B) {
	input := []byte(`{"data":"` + strings.Repeat("x", 1<<20) + `"}`)
	opt := DefaultOptions()
	opt.Pretty = false
	opt.PreserveIfValid = false
	b.SetBytes(int64(len(input)))

	for i := 0; i < b.N; i++ {
		res, err := FixBytes(input, opt)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkResult = res
	}
}

func BenchmarkFixBytesTolerantCommentsAndCommas(b *testing.B) {
	input := []byte(`{// line
"items":[` + strings.Repeat(`{"name":"stone",/* block */"value":1,},`, 1024) + `]}`)
	opt := DefaultOptions()
	opt.Pretty = false
	opt.PreserveIfValid = false
	b.SetBytes(int64(len(input)))

	for i := 0; i < b.N; i++ {
		res, err := FixBytes(input, opt)
		if err != nil {
			b.Fatal(err)
		}
		benchmarkResult = res
	}
}
