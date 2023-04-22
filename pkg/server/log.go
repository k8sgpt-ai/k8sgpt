// Copyright © 2023 K8sGPT.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"bytes"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	buf        *bytes.Buffer
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{
		w,
		http.StatusOK,
		&bytes.Buffer{},
	}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	return lrw.buf.Write(b)
}

func (lrw *loggingResponseWriter) Flush() {
	if f, ok := lrw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
	lrw.ResponseWriter.Write(lrw.buf.Bytes())
}

func logRequest(logger *zap.Logger, fields []zap.Field, statusCode int, message string) {
	if statusCode >= 400 {
		logger.Error(message, fields...)
	} else {
		logger.Info("request completed", fields...)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	config := zap.NewProductionConfig()
	config.DisableCaller = true
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := NewLoggingResponseWriter(w)
		start := time.Now()
		defer func() {
			duration := time.Since(start).Milliseconds()
			fields := []zap.Field{
				zap.Int64("duration_ms", duration),
				zap.String("method", r.Method),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Int("status_code", lrw.statusCode),
				zap.String("url", r.URL.Path),
			}
			logRequest(logger, fields, lrw.statusCode, lrw.buf.String())
		}()

		next.ServeHTTP(lrw, r)

		lrw.Flush()
	})
}
