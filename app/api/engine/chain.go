package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/omegaatt36/bookly/app"
)

func encodeJSON[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

// Empty represents an empty struct.
type Empty struct{}

// Context represents a context.
type Context struct {
	Request *http.Request
}

// Handler collects input and output adapter behavior.
type Handler[Req, Resp any] struct {
	r    *http.Request
	w    http.ResponseWriter
	call func(*Context, Req) (Resp, error)

	resp Resp
	err  error
}

// Chain creates a new handler.
func Chain[Req, Resp any](r *http.Request, w http.ResponseWriter, call func(*Context, Req) (Resp, error)) *Handler[Req, Resp] {
	return &Handler[Req, Resp]{
		r:    r,
		w:    w,
		call: call,
	}
}

// Param sets a parameter.
func (h *Handler[Req, Resp]) Param(key string, val any) *Handler[Req, Resp] {
	if h.err != nil {
		return h
	}

	s := h.r.PathValue(key)
	if s == "" {
		h.err = app.ParamError(fmt.Errorf("parameter '%s' is required", key))
		return h
	}
	if err := condConvert(s, val); err != nil {
		h.err = app.ParamError(fmt.Errorf("convert %v failed: %w", key, err))
		return h
	}

	return h
}

// Query sets a query parameter.
func (h *Handler[Req, Resp]) Query(key string, val any) *Handler[Req, Resp] {
	if h.err != nil {
		return h
	}

	s := h.r.URL.Query().Get(key)
	if s == "" {
		return h
	}

	if err := condConvert(s, val); err != nil {
		h.err = fmt.Errorf("convert %v failed: %w", key, err)
		return h
	}

	return h
}

// BindJSON binds a request body to a struct.
func (h *Handler[Req, Resp]) BindJSON(req *Req) *Handler[Req, Resp] {
	if h.err != nil {
		return h
	}

	if err := json.NewDecoder(h.r.Body).Decode(&req); err != nil {
		h.err = err
		return h
	}

	return h
}

// Call calls the handler.
func (h *Handler[Req, Resp]) Call(req Req) *Handler[Req, Resp] {
	if h.err != nil {
		return h
	}

	h.resp, h.err = h.call(&Context{h.r}, req)

	return h
}

// ResponseJSON encodes the response as JSON.
func (h *Handler[Req, Resp]) ResponseJSON() {
	if h.err != nil {
		h.responseError()
		return
	}

	if err := encodeJSON(h.w, http.StatusOK, Response{Data: h.resp}); err != nil {
		panic(err)
	}
}

// ResponseCreated encodes the response as JSON and sets the status code to 201.
func (h *Handler[Req, Resp]) ResponseCreated() {
	if h.err != nil {
		h.responseError()
		return
	}

	if err := encodeJSON(h.w, http.StatusCreated, Response{Data: h.resp}); err != nil {
		panic(err)
	}
}

func (h *Handler[Req, Resp]) responseError() {
	if h.err == nil {
		slog.ErrorContext(h.r.Context(), "call responseError with nil error")

		if err := encodeJSON(h.w, http.StatusInternalServerError, ResponseError{
			Code:    app.CodeInternalError,
			Message: "",
		}); err != nil {
			panic(err)
		}
		return
	}

	statusCode := http.StatusInternalServerError
	res := ResponseError{
		Code:    app.CodeInternalError,
		Message: h.err.Error(),
	}

	var codeError *app.CodedError
	if errors.As(h.err, &codeError) {
		statusCode = codeError.StatusCode
		res.Code = codeError.AppCode
		res.Message = codeError.Error()
	}

	if err := encodeJSON(h.w, statusCode, res); err != nil {
		panic(err)
	}
}
