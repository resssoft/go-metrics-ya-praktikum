package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
	"github.com/rs/zerolog/log"
)

type PgManager struct {
	storage *sql.DB
}

func New(address string) (structure.Storage, error) {
	//host=%s port=%d user=%s password=%s dbname=%s sslmode=disable
	pgManang, err := sql.Open("postgres", address)
	CheckError(err)
	err = pgManang.Ping()
	CheckError(err)
	log.Info().Msg("postgres Connected!")
	data := &PgManager{
		storage: pgManang,
	}
	data.Init()
	return data, err
}

func (s *PgManager) Close() {
	err := s.storage.Close()
	if err != nil {
		log.Info().Err(err).Msg("Close pg error")
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *PgManager) DB() *sql.DB {
	return s.storage
}

func (s *PgManager) Init() {
	query := "create table IF NOT EXISTS ypt (id VARCHAR not null,mtype VARCHAR not null,delta bigint,value double precision);create unique index IF NOT EXISTS ypt_id_uindex on ypt (id);DO $$ BEGIN IF NOT EXISTS (SELECT FROM ypt limit 1) THEN alter table ypt add constraint ypt_pk primary key (id); END IF; END $$;"
	rows, err := s.DB().Query(query)
	if err != nil {
		log.Info().Err(err).Msg("Init pg error")
		return
	}
	rowsErr := rows.Err()
	if rowsErr != nil {
		log.Info().Err(err).Msg("rowsErr")
	}
	log.Info().Msg("pg init done")
}

func (s *PgManager) Ping() string {
	err := s.DB().Ping()
	if err != nil {
		return err.Error()
	}
	return ""
}

func (s *PgManager) SaveGauge(key string, val models.Gauge) {
	metric, err := s.getByName(key)
	if err != nil || metric.ID == "" {
		log.Debug().AnErr("SaveGauge pg error", err).Msg("SaveGauge pg error")
		s.save(key, "gauge", models.Counter(0), val)
	} else {
		s.update(key, "gauge", models.Counter(0), val)
	}
}

func (s *PgManager) SaveCounter(key string, val models.Counter) {
	_, err := s.getByName(key)
	if err != nil {
		log.Debug().AnErr("SaveGauge pg error", err).Msg("SaveGauge pg error")
		s.save(key, "counter", val, models.Gauge(0))
	} else {
		s.update(key, "counter", val, models.Gauge(0))
	}
}

func (s *PgManager) GetGauges() map[string]models.Gauge {
	result := make(map[string]models.Gauge)
	items, _ := s.getByType("gauge")
	for _, item := range items {
		result[item.ID] = models.Gauge(getDBSafelyValue(item.Value))
	}
	return result
}

func (s *PgManager) GetCounters() map[string]models.Counter {
	result := make(map[string]models.Counter)
	items, _ := s.getByType("counter")
	for _, item := range items {
		result[item.ID] = models.Counter(getDBSafelyDelta(item.Delta))
	}
	return result
}

func (s *PgManager) IncrementCounter(key string, val models.Counter) {
	_, err := s.getByName(key)
	if err == nil {
		val++
		s.update(key, "counter", val, models.Gauge(0)) //+models.Counter(getDBSafelyDelta(metric.Delta))
	} else {
		s.update(key, "counter", val, models.Gauge(0))
	}
}

func (s *PgManager) GetCounter(key string) (models.Counter, error) {
	metric, err := s.getByName(key)
	var value models.Counter
	if err == nil {
		value = models.Counter(getDBSafelyDelta(metric.Delta))
	}
	return value, err
}

func (s *PgManager) GetGauge(key string) (models.Gauge, error) {
	metric, err := s.getByName(key)
	var value models.Gauge
	if err == nil {
		value = models.Gauge(getDBSafelyValue(metric.Value))
	}
	return value, err
}

func (s *PgManager) getByName(name string) (structure.Metrics, error) {
	result := structure.Metrics{}
	query := fmt.Sprintf(`SELECT * FROM ypt where id='%s'`, name)
	log.Debug().Msg("getByName query: " + query)
	rows, err := s.DB().Query(query)
	if err != nil {
		return result, err
	}
	rowsErr := rows.Err()
	if rowsErr != nil {
		return result, rowsErr
	}
	defer rows.Close()
	if rows.Next() {
		var id string
		var mtype string
		var delta *int64
		var value *float64
		err = rows.Scan(&id, &mtype, &delta, &value)
		if err != nil {
			return result, err
		}
		result = structure.Metrics{
			ID:    id,
			MType: mtype,
			Delta: delta,
			Value: value,
		}
	}
	return result, err
}

func (s *PgManager) getByType(mtype string) ([]structure.Metrics, error) {
	var result []structure.Metrics
	query := fmt.Sprintf(`SELECT * FROM ypt where mtype ='%s'`, mtype)
	log.Debug().Msg("getByType query: " + query)
	rows, err := s.DB().Query(query)
	if err != nil {
		return result, err
	}
	rowsErr := rows.Err()
	if rowsErr != nil {
		return result, rowsErr
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var mtype string
		var delta *int64
		var value *float64
		err = rows.Scan(&id, &mtype, &delta, &value)
		if err != nil {
			return result, err
		}
		result = append(result, structure.Metrics{
			ID:    id,
			MType: mtype,
			Delta: delta,
			Value: value,
		})
	}
	return result, err
}

func (s *PgManager) save(id, mtype string, delta models.Counter, value models.Gauge) {
	queryTmp := `insert into "ypt" ("id", "mtype", "delta", "value") values('%s', '%s', %v, %v)`
	query := fmt.Sprintf(queryTmp, id, mtype, delta, value)
	log.Debug().Msg("save query: " + query)
	_, err := s.DB().Exec(query)
	if err != nil {
		log.Info().Err(err).Msg("save item error")
	}
}

func (s *PgManager) update(id, mtype string, delta models.Counter, value models.Gauge) {
	query := `update "ypt" set "id"=$1, "mtype"=$2, "delta"=$3, "value"=$4 where "id"=$5`
	log.Debug().Msg("update query: " + query)
	_, err := s.DB().Exec(query, id, mtype, delta, value, id)
	if err != nil {
		log.Info().Err(err).Msg("update item error")
	}
}

func getDBSafelyDelta(link *int64) int64 {
	if link == nil {
		return 0
	}
	return *link
}

func getDBSafelyValue(link *float64) float64 {
	if link == nil {
		return 0.0
	}
	return *link
}
