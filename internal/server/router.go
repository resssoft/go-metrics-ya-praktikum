package server

import (
	"compress/gzip"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func Router(storage structure.Storage) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(gzipHandle)
	handler := NewMetricsSaver(storage)
	router.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetAll)
		r.Route("/update", func(r chi.Router) {
			r.Post("/gauge/{name}/{value}", handler.SaveGauge)
			r.Post("/counter/{name}/{value}", handler.SaveCounter)
			r.Post("/", handler.SaveValue)
			r.Post("/gauge/{name}/", handler.h404)
			r.Post("/counter/{name}/", handler.h404)
			r.Post("/gauge/", handler.h404)
			r.Post("/counter/", handler.h404)
			r.Post("/*", handler.h501) // it is wrong, i think, but autotests think otherwise
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/gauge/{name}", handler.GetGauge)
			r.Get("/counter/{name}", handler.GetCounter)
			r.Post("/", handler.GetValue)
			r.Get("/gauge/", handler.h404)
			r.Get("/counter/", handler.h404)
			r.Post("/*", handler.h501) // it is wrong, i think, but autotests think otherwise
		})
	})
	return router
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func gzipHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("middleware", r.URL.Path)
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}
