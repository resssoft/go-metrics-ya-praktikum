package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
)

func Router(storage structure.Storage) chi.Router {
	router := chi.NewRouter()
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
