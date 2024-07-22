package main

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/n404an/toolkit/env"
	"github.com/n404an/toolkit/tntV2"
	"log"
	"testing"
)

/*
goos: linux
goarch: amd64
pkg: bitbucket.org/madgpt56/bitsgap-utils
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkTnt/tnt-8                   891           1532983 ns/op          329894 B/op       6852 allocs/op
BenchmarkTnt/tntV2-8              6221655               166.7 ns/op             0 B/op          0 allocs/op
BenchmarkParallelTnt-8           2496746               453.3 ns/op           144 B/op          3 allocs/op
BenchmarkParallelTnt2-8          2218466               534.8 ns/op             0 B/op          0 allocs/op
*/

func TestConnectToTnt2(t *testing.T) {
	env.Load()

	connMap := env.GetString("CONN_MAP", "")
	port := 3220
	proc := "box.space.user_data_subscriptions:select"
	args := []any{}

	tntV2.StopNotify()
	conn := tntV2.NewTntConn()

	connMapStruct := map[int]*tntV2.ConnStruct{}
	if err := jsoniter.Unmarshal([]byte(connMap), &connMapStruct); err != nil {
		log.Println("[ERROR]", err)
		return
	}
	conn.SetGlobalConnMap(connMapStruct)

	tuples, err := conn.CallByPort(port, proc, args)
	if err != nil {
		t.Error(err)
		return
	}
	// fmt.Println(tuples[0])
	if len(tuples) == 0 {
		t.Error("wrong data", tuples)
	}
}

func BenchmarkTntSelect(b *testing.B) {
	env.Load()

	connMap := env.GetString("CONN_MAP", "")
	port := 3220
	proc := "box.space.user_data_subscriptions:select"
	args := []any{}

	b.Run("v2", func(b *testing.B) {
		tntV2.StopNotify()
		conn := tntV2.NewTntConn()

		connMapStruct := map[int]*tntV2.ConnStruct{}
		if err := jsoniter.Unmarshal([]byte(connMap), &connMapStruct); err != nil {
			log.Println("[ERROR]", err)
			return
		}
		conn.SetGlobalConnMap(connMapStruct)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tuples, err := conn.CallByPort(port, proc, args)
			if err != nil {
				b.Error(err)
				return
			}
			// fmt.Println(tuples[0])
			if len(tuples) == 0 {
				b.Error("wrong data", tuples)
			}
		}
	})
}
