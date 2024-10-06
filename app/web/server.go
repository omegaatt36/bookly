package web

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

//go:embed templates/*.html
var templatesFS embed.FS

// Server represents a web server
type Server struct {
	port      int
	router    http.Handler
	templates *template.Template

	serverURL string
}

// NewServer creates a new web server
func NewServer(options ...Option) *Server {
	server := &Server{
		serverURL: "http://localhost:8080",
		port:      3000,
	}

	for _, option := range options {
		option.apply(server)
	}

	server.initTemplates()
	server.registerRoutes()

	return server
}

func (s *Server) initTemplates() {
	funcMap := template.FuncMap{
		"now": func() time.Time {
			return time.Now()
		},
	}

	templates, err := template.New("templates").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		slog.Error("failed to parse templates", slog.String("error", err.Error()))
	}

	s.templates = templates
}

// Run starts the server.
func (s *Server) Run(ctx context.Context) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", slog.String("error", err.Error()))
		}
	}()

	slog.Info("starting web server", slog.String("addr", srv.Addr))

	if err := srv.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", slog.String("error", err.Error()))
	}
}

type sendRequestError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *sendRequestError) Error() string {
	return fmt.Sprintf("failed to send request: %s", e.Message)
}

func (s *Server) sendRequest(r *http.Request, method, path string, body any, result any) error {
	url := fmt.Sprintf("%s%s", s.serverURL, path)
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if !strings.HasPrefix(path, "/public") {
		token, err := r.Cookie("token")
		if err != nil {
			return fmt.Errorf("failed to get token from cookie: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token.Value)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Code    int             `json:"code"`
		Data    json.RawMessage `json:"data"`
		Message string          `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Code != 0 {
		return &sendRequestError{
			Code:    response.Code,
			Message: fmt.Sprintf("failed to send request: %s", response.Message),
		}
	}

	if result == nil {
		return nil
	}

	if err := json.Unmarshal(response.Data, result); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}
