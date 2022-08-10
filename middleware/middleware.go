package middleware

import (
	"github.com/gorilla/mux"
	"net/http"
)

// https://github.com/gorilla/mux

/*
mux一般使用：
	func Middleware(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ...
			next.ServeHTTP(w, r)
		})
	}
	r := mux.NewRouter()
	r.HandleFunc("/", handler)
	r.Use(loggingMiddleware)

maxAllowMiddleware：是返回一个Middleware的函数
maxAllow = maxAllowMiddleware(n)
r.Use(maxAllow)
*/
func MaxAllowMiddleware(n uint) mux.MiddlewareFunc {
	// 通过chan进行简单的限流
	ch := make(chan struct{}, n)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ch <- struct{}{}
			defer func() {
				<-ch
			}()
			next.ServeHTTP(w, r)
		})
	}
}
