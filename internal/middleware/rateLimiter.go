package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

var mutex = sync.RWMutex{}

var rateLimitMap = make(map[string]int)

const rateLimit int = 30

const limitTimeout = time.Second * 8

func RateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		mutex.Lock()
		ip := r.RemoteAddr
		rateLimitMap[ip]++
		mutex.Unlock()

		if rateLimitMap[ip] > rateLimit {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func LimitTimeoutRoutine(ctx context.Context) {
	ticker := time.NewTicker(limitTimeout)
	defer ticker.Stop()

	for {

		select {
		case <-ticker.C:
			mutex.Lock()
			for k, n := range rateLimitMap {
				if n == 0 {
					//fmt.Printf("Deleting ip %s from rateLimitMap\n", k)
					delete(rateLimitMap, k)
				} else if n > 0 {
					rateLimitMap[k]--
					//fmt.Printf("Subtracting from ip: %s = %d\n", k, n)
				}
			}
			mutex.Unlock()
		case <-ctx.Done():
			fmt.Printf("Returning from rateLimit timeout routine...\n")
			return
		}

	}
}
