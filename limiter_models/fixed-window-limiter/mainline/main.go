package main

import (
	"context"
	"fmt"
	limiter_models "github.com/ntrajic/rate-limiters/limiter-models"
	"net/http"
	"time"
)

func main() {
	l := NewFixedWindowLimiter(0, time.Second)
	// l.NewFixedWindowLimiter(10, time.Second*30)
	// l.NewFixedWindowLimiter(15, time.Minute)
	count := 0
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		err := l.TryAcquire(context.Background(), "test")
		if err == nil {
			w.Write([]byte("&&&" + time.Now().String()))
			count++
			fmt.Println(count)
		} else {
			w.Write([]byte(err.Error() + time.Now().String()))
		}
	})
	http.ListenAndServe("127.0.0.1:8080", nil)
}

func NewFixedWindowLimiter(i int, duration time.Duration) {
	panic("unimplemented")
}