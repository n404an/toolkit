package log_test

import (
	"context"
	"github.com/n404an/toolkit/log"
	"sync"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	log.Start(ctx, wg, "INFO")

	go func() {
		log.Debug("test1").Ln(123) // не выведется, т.к. уровень не ниже info
		log.Info("test2").Ln(1e2)
		log.Error("test3").Ln(404)
		time.Sleep(time.Millisecond * 50)
		log.Fatal("test", "4").Ln(false) // не выведется, т.к. логгер уже не работает
	}()
	time.Sleep(time.Millisecond * 40)
	cancel()
	wg.Wait()

	log.Info("test").Ln("done") // не выведется, т.к. логгер уже не работает
}
