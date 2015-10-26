package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang/glog"
)

const (
	logFmt = "%s \"%s %d %d\" %f"
)

type logRecord struct {
	http.ResponseWriter

	ip                    string
	method, uri, protocol string
	status                int
	responseBytes         int64
	elapsedTime           time.Duration
}

func (r *logRecord) Log() {
	requestLine := fmt.Sprintf("%s %s %s", r.method, r.uri, r.protocol)
	glog.Infof(logFmt, r.ip, requestLine, r.status, r.responseBytes, r.elapsedTime.Seconds())
	glog.Flush()
}

func (r *logRecord) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *logRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type loggingHandler struct {
	handler http.Handler
}

func NewLoggingHandler(handler http.Handler) http.Handler {
	return &loggingHandler{
		handler: handler,
	}
}

func (h *loggingHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	record := &logRecord{
		ResponseWriter: rw,
		ip:             clientIP,
		method:         r.Method,
		uri:            r.RequestURI,
		protocol:       r.Proto,
		status:         http.StatusOK,
		elapsedTime:    time.Duration(0),
	}

	h.handler.ServeHTTP(record, r)

	record.elapsedTime = time.Now().Sub(startTime)
	record.Log()
}
