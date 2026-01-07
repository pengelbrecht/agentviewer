// Go Code Test File - Tests syntax highlighting for Go language features
// This file demonstrates various Go language constructs for testing code rendering

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Constants and iota
const (
	StatusPending Status = iota
	StatusRunning
	StatusComplete
	StatusFailed
)

const MaxRetries = 3
const Timeout = 30 * time.Second

// Type definitions
type Status int

type Config struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Debug    bool              `json:"debug,omitempty"`
	Tags     []string          `json:"tags"`
	Settings map[string]string `json:"settings"`
}

// Generic types (Go 1.18+)
type Result[T any] struct {
	Value T
	Error error
}

type Cache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

// Interface definition
type Handler interface {
	Handle(ctx context.Context, req *Request) (*Response, error)
	Close() error
}

type Request struct {
	ID      string
	Payload []byte
}

type Response struct {
	Status  int
	Body    []byte
	Headers http.Header
}

// Struct with embedded types and methods
type Server struct {
	config  *Config
	handler Handler
	mu      sync.RWMutex
	done    chan struct{}
}

// Constructor pattern
func NewServer(cfg *Config, h Handler) *Server {
	return &Server{
		config:  cfg,
		handler: h,
		done:    make(chan struct{}),
	}
}

// Method with pointer receiver
func (s *Server) Start(ctx context.Context) error {
	// Goroutine and channel usage
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-s.done:
				return
			case t := <-ticker.C:
				fmt.Printf("Tick at %v\n", t)
			}
		}
	}()

	return nil
}

// Generic function
func Map[T, U any](items []T, fn func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}
	return result
}

// Variadic function
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Error handling patterns
var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrTimeout      = errors.New("operation timed out")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// Function with multiple return values
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// Defer, panic, recover
func SafeOperation(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic: %v", r)
		}
	}()

	fn()
	return nil
}

// Type switch
func TypeCheck(v interface{}) string {
	switch t := v.(type) {
	case int:
		return fmt.Sprintf("int: %d", t)
	case string:
		return fmt.Sprintf("string: %s", t)
	case bool:
		return fmt.Sprintf("bool: %t", t)
	case []byte:
		return fmt.Sprintf("bytes: %d bytes", len(t))
	case nil:
		return "nil"
	default:
		return fmt.Sprintf("unknown type: %T", t)
	}
}

// Closures
func Counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// Working with JSON
func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

// HTTP handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Context with timeout
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	req := &Request{
		ID:      r.Header.Get("X-Request-ID"),
		Payload: body,
	}

	resp, err := s.handler.Handle(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for key, values := range resp.Headers {
		for _, v := range values {
			w.Header().Add(key, v)
		}
	}

	w.WriteHeader(resp.Status)
	w.Write(resp.Body)
}

// Main function
func main() {
	// String literals
	raw := `This is a
	raw string with "quotes" and \backslashes`

	multiline := "Line 1\n" +
		"Line 2\n" +
		"Line 3"

	// Slice and map literals
	numbers := []int{1, 2, 3, 4, 5}
	lookup := map[string]int{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	// Range over slice
	for i, n := range numbers {
		fmt.Printf("Index %d: %d\n", i, n)
	}

	// Range over map
	for key, value := range lookup {
		fmt.Printf("Key %s: %d\n", key, value)
	}

	// Type assertion
	var x interface{} = "hello"
	if s, ok := x.(string); ok {
		fmt.Println("String value:", s)
	}

	// Anonymous struct
	person := struct {
		Name string
		Age  int
	}{
		Name: "Alice",
		Age:  30,
	}

	fmt.Printf("Person: %+v\n", person)
	fmt.Println(raw, multiline, numbers, lookup)
}
