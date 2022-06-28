package brotli

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/andybalholm/brotli"
)

func Call(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		w := httptest.NewRecorder()
		for name, vv := range rw.Header() {
			for _, v := range vv {
				w.Header().Add(name, v)
			}
		}

		next.ServeHTTP(w, req)

		for name := range rw.Header() {
			rw.Header().Del(name)
		}
		for name, vv := range w.Header() {
			for _, v := range vv {
				rw.Header().Add(name, v)
			}
		}
		rw.Header().Del("Content-Length")
		rw.Header().Del("Content-Encoding")

		buf := bytes.NewBuffer(nil)

		brWriter := brotli.NewWriter(buf)

		_, errWrite := brWriter.Write(w.Body.Bytes())
		if errWrite != nil {
			log.Printf("[ERROR] error write, %v", errWrite)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		errFlush := brWriter.Flush()
		if errFlush != nil {
			log.Printf("[ERROR] error flush, %v", errFlush)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		errClose := brWriter.Close()
		if errClose != nil {
			log.Printf("[ERROR] error close, %v", errClose)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Encoding", "br")
		rw.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
		rw.WriteHeader(w.Code)
		_, errWriteResp := rw.Write(buf.Bytes())
		if errWriteResp != nil {
			log.Printf("[ERROR] error write response, %v", errWriteResp)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
