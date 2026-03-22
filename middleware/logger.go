package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"one-api/common/config"
	"one-api/common/logger"
	"one-api/metrics"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/gin-gonic/gin"
)

const maxLoggedBodySize = 8 * 1024

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(data []byte) (int, error) {
	w.appendBody(data)
	return w.ResponseWriter.Write(data)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.appendBody([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

func (w *bodyLogWriter) appendBody(data []byte) {
	if w.body == nil || len(data) == 0 {
		return
	}
	remaining := maxLoggedBodySize - w.body.Len()
	if remaining <= 0 {
		return
	}
	if len(data) > remaining {
		_, _ = w.body.Write(data[:remaining])
		return
	}
	_, _ = w.body.Write(data)
}

func SetUpLogger(server *gin.Engine) {
	server.Use(GinzapWithConfig())
}

func GinzapWithConfig() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		requestBody, requestBodyTruncated := captureRequestBody(c)
		blw := &bodyLogWriter{ResponseWriter: c.Writer, body: bytes.NewBuffer(make([]byte, 0, maxLoggedBodySize))}
		c.Writer = blw

		c.Next()

		latency := time.Since(start)
		requestID := c.GetString(logger.RequestIdKey)
		userID := c.GetInt("id")
		statusCode := c.Writer.Status()
		responseBody := ""
		responseBodyTruncated := false
		if shouldLogBody(c.Writer.Header().Get("Content-Type")) {
			responseBody = sanitizeText(blw.body.String())
			responseBodyTruncated = blw.body.Len() >= maxLoggedBodySize
		}

		fields := []zapcore.Field{
			zap.Int("status", statusCode),
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", sanitizeQuery(c.Request.URL.RawQuery)),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Duration("latency", latency),
			zap.Int("user_id", userID),
			zap.String("username", c.GetString("username")),
			zap.String("original_model", c.GetString("original_model")),
			zap.String("new_model", c.GetString("new_model")),
			zap.Int("token_id", c.GetInt("token_id")),
			zap.String("token_name", c.GetString("token_name")),
			zap.Int("channel_id", c.GetInt("channel_id")),
			zap.Int("channel_type", c.GetInt("channel_type")),
		}

		if len(c.Errors) > 0 || statusCode >= http.StatusBadRequest {
			fields = append(fields,
				zap.String("request_headers", sanitizeHeaders(c.Request.Header)),
				zap.String("request_body", requestBody),
				zap.Bool("request_body_truncated", requestBodyTruncated),
				zap.String("response_headers", sanitizeHeaders(c.Writer.Header())),
				zap.String("response_body", responseBody),
				zap.Bool("response_body_truncated", responseBodyTruncated),
			)
			if len(c.Errors) > 0 {
				fields = append(fields, zap.Strings("gin_errors", c.Errors.Errors()))
			}
			logger.Logger.Error("GIN error request", fields...)
		} else {
			logger.Logger.Info("GIN request", fields...)
		}
		metrics.RecordHttp(c, latency)
	}
}

func captureRequestBody(c *gin.Context) (string, bool) {
	if raw, exists := c.Get(config.GinRequestBodyKey); exists {
		if body, ok := raw.([]byte); ok {
			return sanitizeBytes(body)
		}
	}

	if c.Request == nil || c.Request.Body == nil || !shouldLogBody(c.ContentType()) {
		return "", false
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Sprintf("[read body failed: %v]", err), false
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return sanitizeBytes(body)
}

func sanitizeBytes(body []byte) (string, bool) {
	truncated := false
	if len(body) > maxLoggedBodySize {
		body = body[:maxLoggedBodySize]
		truncated = true
	}
	return sanitizeText(string(body)), truncated
}

func sanitizeQuery(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return sanitizeText(rawQuery)
	}
	for key := range values {
		if isSensitiveKey(key) {
			values[key] = []string{"***"}
			continue
		}
		for idx, value := range values[key] {
			values[key][idx] = maskSensitiveValue(key, value)
		}
	}
	return sanitizeText(values.Encode())
}

func sanitizeHeaders(headers http.Header) string {
	if len(headers) == 0 {
		return ""
	}
	masked := make(map[string][]string, len(headers))
	for key, values := range headers {
		copied := make([]string, len(values))
		for idx, value := range values {
			if isSensitiveKey(key) {
				copied[idx] = "***"
			} else {
				copied[idx] = maskSensitiveValue(key, value)
			}
		}
		masked[key] = copied
	}
	keys := make([]string, 0, len(masked))
	for key := range masked {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	builder := strings.Builder{}
	for idx, key := range keys {
		if idx > 0 {
			builder.WriteString("; ")
		}
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(strings.Join(masked[key], ","))
	}
	return sanitizeText(builder.String())
}

func sanitizeText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if sanitized, ok := sanitizeJSON(text); ok {
		return sanitized
	}
	return sanitizePlainText(text)
}

func sanitizeJSON(text string) (string, bool) {
	var data interface{}
	decoder := json.NewDecoder(strings.NewReader(text))
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return "", false
	}
	sanitizeJSONValue(&data, "")
	body, err := json.Marshal(data)
	if err != nil {
		return "", false
	}
	return string(body), false
}

func sanitizeJSONValue(value *interface{}, key string) {
	switch typed := (*value).(type) {
	case map[string]interface{}:
		for childKey, childValue := range typed {
			if isSensitiveKey(childKey) {
				typed[childKey] = "***"
				continue
			}
			copied := childValue
			sanitizeJSONValue(&copied, childKey)
			typed[childKey] = copied
		}
	case []interface{}:
		for idx := range typed {
			copied := typed[idx]
			sanitizeJSONValue(&copied, key)
			typed[idx] = copied
		}
	case string:
		if isSensitiveKey(key) {
			*value = "***"
			return
		}
		*value = maskSensitiveValue(key, typed)
	}
}

func sanitizePlainText(text string) string {
	masked := text
	patterns := []string{
		"Bearer ",
		"sk-",
	}
	for _, pattern := range patterns {
		for {
			idx := strings.Index(masked, pattern)
			if idx < 0 {
				break
			}
			end := idx + len(pattern)
			for end < len(masked) {
				ch := masked[end]
				if ch == ' ' || ch == '\n' || ch == '\r' || ch == '\t' || ch == '"' || ch == '\'' || ch == ',' || ch == '&' {
					break
				}
				end++
			}
			masked = masked[:idx+len(pattern)] + "***" + masked[end:]
		}
	}
	return masked
}

func maskSensitiveValue(key string, value string) string {
	if value == "" {
		return value
	}
	if isSensitiveKey(key) {
		return "***"
	}
	lowerValue := strings.ToLower(value)
	if strings.HasPrefix(lowerValue, "bearer ") {
		return "Bearer ***"
	}
	if strings.HasPrefix(value, "sk-") {
		return "sk-***"
	}
	return value
}

func isSensitiveKey(key string) bool {
	key = strings.ToLower(key)
	sensitiveKeywords := []string{
		"authorization",
		"api-key",
		"apikey",
		"x-api-key",
		"x-goog-api-key",
		"mj-api-secret",
		"token",
		"secret",
		"password",
		"key",
	}
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(key, keyword) {
			return true
		}
	}
	return false
}

func shouldLogBody(contentType string) bool {
	if contentType == "" {
		return true
	}
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return strings.HasPrefix(mediaType, "application/json") ||
		strings.HasPrefix(mediaType, "text/") ||
		mediaType == "application/x-www-form-urlencoded"
}
