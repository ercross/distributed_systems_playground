package httpx

import "net/http"

type ResponseRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.status = statusCode
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(b)
}
