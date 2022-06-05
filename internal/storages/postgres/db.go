package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
)

type DbData struct {
	storage *sql.DB
}

func New(address string) (structure.Storage, error) {
	//host=%s port=%d user=%s password=%s dbname=%s sslmode=disable
	db, err := sql.Open("postgres", address)
	CheckError(err)
	defer db.Close()
	err = db.Ping()
	CheckError(err)
	fmt.Println("Connected!")
	return &DbData{
		storage: db,
	}, err
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *DbData) Ping() string {
	err := s.storage.Ping()
	if err != nil {
		return err.Error()
	}
	return ""
}

func (s *DbData) SaveGauge(key string, val models.Gauge) {
}

func (s *DbData) SaveCounter(key string, val models.Counter) {
}

func (s *DbData) GetGauges() map[string]models.Gauge {
	result := make(map[string]models.Gauge)
	return result
}

func (s *DbData) GetCounters() map[string]models.Counter {
	result := make(map[string]models.Counter)
	return result
}

func (s *DbData) IncrementCounter(key string, val models.Counter) {
}

func (s *DbData) GetCounter(key string) (models.Counter, error) {
	var err error = nil
	var value models.Counter
	return value, err
}

func (s *DbData) GetGauge(key string) (models.Gauge, error) {
	var err error = nil
	var value models.Gauge
	return value, err
}
