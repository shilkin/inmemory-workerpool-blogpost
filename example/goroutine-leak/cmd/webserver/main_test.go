package main

import (
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/analytics"
	pool "github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/pool"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/repo"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/delay"
	"github.com/shilkin/inmemory-workerpool-blogpost/example/goroutine-leak/cmd/webserver/internal/service"

	"github.com/stretchr/testify/require"
)

func setupTestServer(userService *service.UserService) *http.ServeMux {
	server := http.NewServeMux()

	server.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	server.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		name := "John Doe"
		email := "john.doe@example.com"
		d, err := time.ParseDuration(q.Get("delay"))
		if err != nil {
			http.Error(w, "invalid delay parameter", http.StatusBadRequest)
			return
		}

		ctx := delay.ToContext(r.Context(), d)

		if err := userService.Create(ctx, name, email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("User created successfully"))
	})

	return server
}

type ServerSuite struct {
	suite.Suite
	userService *service.UserService
	server      *http.ServeMux
}

func (s *ServerSuite) SetupTest() {
	workerPool := pool.NewPoolService(100)
	s.userService = service.NewUserService(
		repo.NewUserRepo(),
		analytics.NewAnalytics(),
		workerPool,
	)
	s.server = setupTestServer(s.userService)
}

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerSuite))
}

func TestPingEndpoint(t *testing.T) {
	// Создаем userService перед каждым тестом
	workerPool := pool.NewPoolService(100)
	userService := service.NewUserService(
		repo.NewUserRepo(),
		analytics.NewAnalytics(),
		workerPool,
	)

	server := setupTestServer(userService)
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Equal(t, "pong", resp.Body.String())
}

func TestCreateEndpoint_Success(t *testing.T) {
	// Создаем userService перед каждым тестом
	workerPool := pool.NewPoolService(100)
	userService := service.NewUserService(
		repo.NewUserRepo(),
		analytics.NewAnalytics(),
		workerPool,
	)

	server := setupTestServer(userService)
	req := httptest.NewRequest(http.MethodGet, "/create?delay=100ms", nil)
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	require.Contains(t, resp.Body.String(), "User created successfully")
}

func TestCreateEndpoint_InvalidDelay(t *testing.T) {
	workerPool := pool.NewPoolService(100)
	userService := service.NewUserService(
		repo.NewUserRepo(),
		analytics.NewAnalytics(),
		workerPool,
	)

	server := setupTestServer(userService)
	req := httptest.NewRequest(http.MethodGet, "/create?delay=invalid", nil)
	resp := httptest.NewRecorder()

	server.ServeHTTP(resp, req)

	require.Equal(t, http.StatusBadRequest, resp.Code)
	require.Contains(t, resp.Body.String(), "invalid delay parameter")
}
