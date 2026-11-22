package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/bonarizki/gin-boilerplate/pkg/config"
	"github.com/bonarizki/gin-boilerplate/pkg/logger"
	"github.com/bonarizki/gin-boilerplate/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// responseRecorder wraps gin.ResponseWriter to capture status, size, and body for logging.
type responseRecorder struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
	size   int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	r.body.Write(b)
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func shouldLogRequestBodies() bool {
	if config.IsProduction() {
		return config.GetBool("LOG_REQUEST_BODIES")
	}
	if viper.IsSet("LOG_REQUEST_BODIES") {
		return config.GetBool("LOG_REQUEST_BODIES")
	}
	return true
}

func logBodyMaxBytes() int {
	max := config.GetInt("LOG_BODY_LOG_MAX_BYTES")
	if max <= 0 {
		max = 4096
	}
	return max
}

func formatLoggedBody(body []byte, contentType string) string {
	if len(body) == 0 {
		return "{}"
	}
	var masked string
	if json.Valid(body) {
		masked = utils.MaskSensitiveJSON(string(body))
	} else {
		masked = utils.FormatNonJSONBodyLog(contentType, len(body))
	}
	return utils.TruncateForLog(masked, logBodyMaxBytes())
}

// RequestLogMiddleware logs ARRIVED REQUEST (with masked body/query) and RESPONSE SENT (with masked body).
// Must run after RequestIDMiddleware so request_id is available.
func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := c.GetString("request_id")
		if requestID == "" {
			c.Next()
			return
		}
		reqLog := logger.WithRequestID(requestID)
		logBodies := shouldLogRequestBodies()

		var bodyBytes []byte
		if logBodies && c.Request.Body != nil {
			var readErr error
			bodyBytes, readErr = io.ReadAll(c.Request.Body)
			if readErr != nil {
				logger.Warnf("failed to read request body for logging: %v", readErr)
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		queryStr := c.Request.URL.RawQuery
		maskedQuery := utils.MaskQueryString(queryStr)

		if logBodies {
			bodyForLog := formatLoggedBody(bodyBytes, c.GetHeader("Content-Type"))
			reqLog.Infof("ARRIVED REQUEST from IP=%s method=%s path=%s query=%s body=%s",
				c.ClientIP(), c.Request.Method, c.Request.URL.Path, maskedQuery, bodyForLog)
		} else {
			reqLog.Infof("ARRIVED REQUEST from IP=%s method=%s path=%s query=%s body=<omitted>",
				c.ClientIP(), c.Request.Method, c.Request.URL.Path, maskedQuery)
		}

		rec := &responseRecorder{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
			status:         http.StatusOK,
		}
		c.Writer = rec

		c.Next()

		durMs := time.Since(start).Seconds() * 1000
		if logBodies {
			maskedRespBody := formatLoggedBody(rec.body.Bytes(), rec.Header().Get("Content-Type"))
			reqLog.Infof("RESPONSE SENT status=%d duration=%.2fms size=%d bytes body=%s",
				rec.status, durMs, rec.size, maskedRespBody)
		} else {
			reqLog.Infof("RESPONSE SENT status=%d duration=%.2fms size=%d bytes body=<omitted>",
				rec.status, durMs, rec.size)
		}
	}
}