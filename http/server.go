package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/mux"
)

const (
	shutdownTimeout  = 10 * time.Second
	readWriteTimeout = 15 * time.Second

	defaultAddr = ":8080"
)

// Server represents http server.
type Server struct {
	srv    *http.Server
	logger logger
}

// Run starts serving and listening http server with graceful shutdown.
func (s *Server) Run() error {
	s.logger.Println("http server: running on", s.srv.Addr)
	return s.gracefulShutdown(s.srv, shutdownTimeout)
}

func (s *Server) gracefulShutdown(srv *http.Server, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c

		ctx := context.Background()
		var cancel context.CancelFunc
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		s.logger.Println("http server: shutdown...")
		done <- srv.Shutdown(ctx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	s.logger.Println("http server: shutdown gracefully")
	return <-done
}

// NewServer creates new instance of http server.
func NewServer(addr string, rest *RestHandler, l logger) *Server {
	if addr == "" {
		addr = defaultAddr
	}

	r := mux.NewRouter()
	r.Use(structuredLogger(l))
	r.Handle("/users", rest.GETUsers()).Methods(http.MethodGet)

	s := &Server{
		srv: &http.Server{
			Handler:      r,
			Addr:         addr,
			WriteTimeout: readWriteTimeout,
			ReadTimeout:  readWriteTimeout,
		},
		logger: l,
	}
	return s
}

type logger interface {
	Print(v ...interface{})
	Println(v ...interface{})
}

func structuredLogger(l logger) func(next http.Handler) http.Handler {
	return chimw.RequestLogger(&chimw.DefaultLogFormatter{Logger: l})
}
