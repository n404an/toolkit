package log

import (
	"log"
	"testing"
)

/*
goos: windows
goarch: amd64
cpu: AMD Ryzen 9 5950X 16-Core Processor
StringToBytes
std-32            355551235   3.239 ns/op    0 B/op   0 allocs/op
unsafe(1.20)-32  1000000000   0.2845 ns/op   0 B/op   0 allocs/op
unsafe-32        1000000000   0.2109 ns/op   0 B/op   0 allocs/op
PASS

BytesToString
std-32             336324378  3.233 ns/op   0 B/op  0 allocs/op
unsafe(1.20)-32   1000000000  0.4269ns/op   0 B/op  0 allocs/op
unsafe-32         1000000000  0.2140ns/op   0 B/op  0 allocs/op
PASS
*/

func TestStringToByte(t *testing.T) {
	str := "asdasodasodhlahsdlahsld"
	log.Printf("%#v\n", []byte(str))
	//log.Printf("%#v\n", sToB(str))
	log.Printf("%#v\n", s2b(str))

	bts := []byte(str)
	log.Printf("%#v\n", string(bts))
	//log.Printf("%#v\n", bToS(bts))
	log.Printf("%#v\n", b2s(bts))
}

func BenchmarkStringToBytes(b *testing.B) {
	str := "asdasodasodhlahsdlahsld"

	b.Run("std", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = []byte(str)
		}
	})
	b.Run("unsafe(1.20)", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = s2b(str)
		}
	})
	/*b.Run("unsafe", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = sToB(str)
		}
	})*/

}

func BenchmarkBytesToString(b *testing.B) {
	bts := []byte("asdasodasodhlahsdlahsld")

	b.Run("std", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = string(bts)
		}
	})
	b.Run("unsafe(1.20)", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = b2s(bts)
		}
	})
	/*b.Run("unsafe", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = bToS(bts)
		}
	})*/
}
