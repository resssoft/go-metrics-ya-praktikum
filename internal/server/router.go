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
			r.Post("/guage/{name}/{value}", handler.SaveGuage)
			r.Post("/counter/{name}/{value}", handler.SaveCounter)
			r.Post("/", handler.SaveValue)
		})
		r.Route("/value", func(r chi.Router) {
			r.Get("/guage/{name}", handler.GetGuage)
			r.Get("/counter/{name}", handler.GetCounter)
			r.Post("/", handler.GetValue)
		})
	})
	return router
}
