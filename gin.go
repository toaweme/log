package log

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ResponseRecorder is used to capture the response body and status
type ResponseRecorder struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

// Write records the response body and writes to the underlying writer
func (r *ResponseRecorder) Write(data []byte) (int, error) {
	r.body.Write(data)
	return r.ResponseWriter.Write(data)
}

// WriteHeader captures the response status code
func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// SlogMiddleware creates a Gin middleware for logging using slog
func SlogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture request details
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		clientIP := c.ClientIP()

		// Clone and read the request body (if needed for logging)
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestBody = string(bodyBytes)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore the body
		}

		// Set up response recording
		recorder := &ResponseRecorder{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer([]byte{}),
			status:         http.StatusOK,
		}
		c.Writer = recorder

		// Process the request
		c.Next()

		// Calculate response time
		latency := time.Since(start)

		// Log request and response details
		Ctx(c).Info("api",
			"method", method,
			"path", path,
			"query", query,
			"client_ip", clientIP,
			"request_body", requestBody,
			"response_body", recorder.body.String(),
			"status", recorder.status,
			"latency", latency,
			"error", c.Errors.ByType(gin.ErrorTypePrivate).String(),
		)
	}
}

func Ctx(reqCtx context.Context) *slog.Logger {
	return Logger.With("id", reqCtx.Value("id"), "agent", reqCtx.Value("client"))
}

const ClientIDHeaderName = "X-Client-ID"
const ClientAgentHeaderName = "X-Client-Agent"

func GinLogTraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Request.Header.Get(ClientIDHeaderName)
		if id == "" {
			id = ID()
		}
		c.Set("id", id)

		client := c.Request.Header.Get(ClientAgentHeaderName)
		if client == "" {
			client = "unknown"
		}
		c.Set("client", client)
	}
}

func ID() string {
	return uuid.New().String()
}
