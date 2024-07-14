package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type SumRequest struct {
	A int `json:"a"`
	B int `json:"b"`
}

type SumResponse struct {
	Result int `json:"result"`
}

// Prometheus Counter
var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_counter",
			Help: "Number of HTTP requests processed, labeled by status code, method, and path.",
		},
		[]string{"code", "method", "path"},
	)
)

func sumHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	var data SumRequest
	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	result := data.A + data.B
	res := SumResponse{Result: result}
	json.NewEncoder(w).Encode(res)

}

func LoggingMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s Request for %s\n", r.Method, r.RequestURI)

		rr := &responseRecoder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rr, r)

		requestCounter.With(prometheus.Labels{"code": strconv.Itoa(rr.statusCode), "method": r.Method, "path": r.RequestURI}).Inc()

	})
}

type responseRecoder struct {
	http.ResponseWriter
	statusCode int
}

func (rr *responseRecoder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func init() {
	prometheus.MustRegister(requestCounter)
}

func main() {
	r := mux.NewRouter()
	r.Use(LoggingMiddleware)

	r.HandleFunc("/sum", sumHandler).Methods("POST")
	r.Handle("/metrics", promhttp.Handler())

	fmt.Println("Server is running at port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))

}
