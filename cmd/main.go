package main

import (
	"fmt"
	"net/http"
	"net"
	"log"
)

var limits map[string]int

func getHello(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)
	_, ok := limits[ip]
	
	if !ok {
		limits[ip] = 10
	}
	
	fmt.Printf("Rate limit: %d\n", limits[ip])
	
	isRateLimited := limits[ip] == 0
	
	if isRateLimited {
		w.Header().Set("Retry-After", "60")
		fmt.Println("Rate Limited")
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		return
	}
	
	limits[ip] -= 1
	w.WriteHeader(http.StatusOK)
	fmt.Println("Hello")
}

func getIP(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	fmt.Println(host)
	return host
}
func main() {
	limits = make(map[string]int)
	http.HandleFunc("/hello", getHello)
	
	fmt.Println("Sever listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
