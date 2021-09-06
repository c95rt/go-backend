package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/parqueoasis/backend/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type ResponseWriter struct {
	Writer http.ResponseWriter
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		Writer: w,
	}
}

type generalResponse struct {
	Errors  []*errorResponse `json:"errors"`
	Success bool             `json:"success"`
	Data    interface{}      `json:"data"`
}

type errorResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Scope   string      `json:"scope"`
	Type    int         `json:"type"`
	Data    interface{} `json:"data"`
}

type ErrOption func(*errorResponse)

func WithErrorType(errType int) ErrOption {
	return func(err *errorResponse) {
		err.Type = errType
	}
}

func WithErrorScope(scope string) ErrOption {
	return func(err *errorResponse) {
		err.Scope = scope
	}
}

func (r *ResponseWriter) writeJSONResponse(code int, errors []*errorResponse, data interface{}) {
	r.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	response := &generalResponse{Errors: errors, Success: errors == nil, Data: data}
	b, err := json.Marshal(response)
	if err != nil {
		r.Writer.WriteHeader(http.StatusInternalServerError)
		r.Writer.Write([]byte(fmt.Sprintf("unexpected error: %v", err)))
	}
	r.Writer.WriteHeader(code)
	if code, err := r.Writer.Write(b); err != nil {
		fmt.Sprintf("could not response - code: %d", code)
	}
}

func (r *ResponseWriter) writePlainJSONResponse(statusCode int, data interface{}) {
	b, err := json.Marshal(data)
	if err != nil {
		r.Writer.WriteHeader(http.StatusInternalServerError)
		r.Writer.Write([]byte(fmt.Sprintf("unexpected error: %v", err)))
	}

	r.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	r.Writer.WriteHeader(statusCode)

	if code, err := r.Writer.Write(b); err != nil {
		fmt.Sprintf("could not response - code: %d", code)
	}
}

func (r *ResponseWriter) WriteJSON(statusCode int, data interface{}, err error, message string) {
	logger := config.GetLogger()
	fields := make(log.Fields)
	fields["status_code"] = statusCode
	if statusCode >= 200 && statusCode <= 299 {
		logger.WithFields(fields).Info("success")
	}
	if statusCode >= 300 {
		if data == nil {
			data = map[string]interface{}{
				"error": message,
			}
		}
		if err == nil {
			err = errors.Errorf(message)
		}
		fields["errors"] = data
		logger.WithFields(fields).Error(err)
	}
	r.writePlainJSONResponse(statusCode, data)
}

func (r *ResponseWriter) JSON(code int, data interface{}) {
	r.writeJSONResponse(code, nil, data)
}

func (r *ResponseWriter) Stringf(code int, format string, args ...interface{}) {
	r.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	r.Writer.WriteHeader(code)
	if code, err := r.Writer.Write([]byte(fmt.Sprintf(format, args...))); err != nil {
		fmt.Sprintf("could not response - code: %d", code)
	}
}

func (r *ResponseWriter) Errorf(code int, format string, args ...interface{}) {
	errors := []*errorResponse{
		{Code: code, Message: fmt.Sprintf(format, args...)},
	}
	r.writeJSONResponse(code, errors, nil)
}

func (r *ResponseWriter) ErrorJ(code int, format string, data interface{}) {
	errors := []*errorResponse{
		{Code: code, Message: format, Data: data},
	}
	r.writeJSONResponse(code, errors, nil)
}

func (r *ResponseWriter) String(code int, msg string) {
	r.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	r.Writer.WriteHeader(code)
	if code, err := r.Writer.Write([]byte(msg)); err != nil {
		fmt.Sprintf("could not response - code: %d", code)
	}
}

func (r *ResponseWriter) Error(code int, msg string, opts ...ErrOption) {
	err := &errorResponse{Code: code, Message: msg}
	for _, With := range opts {
		With(err)
	}
	r.writeJSONResponse(code, []*errorResponse{err}, nil)
}
