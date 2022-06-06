package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/models"
	"github.com/resssoft/go-metrics-ya-praktikum/internal/structure"
	"github.com/rs/zerolog/log"
)

type DbData struct {
	storage *sql.DB
}

func New(address string) (structure.Storage, error) {
	//host=%s port=%d user=%s password=%s dbname=%s sslmode=disable
	db, err := sql.Open("postgres", address)
	CheckError(err)
	err = db.Ping()
	CheckError(err)
	log.Info().Msg("db Connected!")
	dbData := &DbData{
		storage: db,
	}
	dbData.Init()
	return dbData, err
}

func (s *DbData) Close() {
	err := s.storage.Close()
	if err != nil {
		log.Info().Err(err).Msg("Close db error")
	}
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *DbData) Db() *sql.DB {
	return s.storage
}

func (s *DbData) Init() {
	query := "create table IF NOT EXISTS ypt (id VARCHAR not null,mtype VARCHAR not null,delta bigint,value double precision);create unique index IF NOT EXISTS ypt_id_uindex on ypt (id);DO $$ BEGIN IF NOT EXISTS (SELECT FROM ypt limit 1) THEN alter table ypt add constraint ypt_pk primary key (id); END IF; END $$;"
	_, err := s.Db().Query(query)
	if err != nil {
		log.Info().Err(err).Msg("Init db error")
		return
	}
	log.Info().Msg("db init done")
}

func (s *DbData) Ping() string {
	err := s.Db().Ping()
	if err != nil {
		return err.Error()
	}
	return ""
}

func (s *DbData) SaveGauge(key string, val models.Gauge) {
	metric, err := s.getByName(key)
	if err != nil || metric.ID == "" {
		log.Debug().AnErr("SaveGauge db error", err).Msg("SaveGauge db error")
		s.save(key, "gauge", models.Counter(0), val)
	} else {
		s.update(key, "gauge", models.Counter(0), val)
	}
}

func (s *DbData) SaveCounter(key string, val models.Counter) {
	_, err := s.getByName(key)
	if err != nil {
		log.Debug().AnErr("SaveGauge db error", err).Msg("SaveGauge db error")
		s.save(key, "counter", val, models.Gauge(0))
	} else {
		s.update(key, "counter", val, models.Gauge(0))
	}
}

func (s *DbData) GetGauges() map[string]models.Gauge {
	result := make(map[string]models.Gauge)
	items, _ := s.getByType("gauge")
	for _, item := range items {
		result[item.ID] = models.Gauge(getDbSafelyValue(item.Value))
	}
	return result
}

func (s *DbData) GetCounters() map[string]models.Counter {
	result := make(map[string]models.Counter)
	items, _ := s.getByType("counter")
	for _, item := range items {
		result[item.ID] = models.Counter(getDbSafelyDelta(item.Delta))
	}
	return result
}

func (s *DbData) IncrementCounter(key string, val models.Counter) {
	metric, err := s.getByName(key)
	if err == nil {
		s.update(key, "counter", val+models.Counter(getDbSafelyDelta(metric.Delta)), models.Gauge(0))
	}
}

func (s *DbData) GetCounter(key string) (models.Counter, error) {
	metric, err := s.getByName(key)
	var value models.Counter
	if err == nil {
		value = models.Counter(getDbSafelyDelta(metric.Delta))
	}
	return value, err
}

func (s *DbData) GetGauge(key string) (models.Gauge, error) {
	metric, err := s.getByName(key)
	var value models.Gauge
	if err == nil {
		value = models.Gauge(getDbSafelyValue(metric.Value))
	}
	return value, err
}

func (s *DbData) getByName(name string) (structure.Metrics, error) {
	result := structure.Metrics{}
	query := fmt.Sprintf(`SELECT * FROM ypt where id='%s'`, name)
	log.Debug().Msg("getByName query: " + query)
	rows, err := s.Db().Query(query)
	if err != nil {
		return result, err
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

func (s *DbData) getByType(mtype string) ([]structure.Metrics, error) {
	var result []structure.Metrics
	query := fmt.Sprintf(`SELECT * FROM ypt where mtype ='%s'`, mtype)
	log.Debug().Msg("getByType query: " + query)
	rows, err := s.Db().Query(query)
	if err != nil {
		return result, err
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

func (s *DbData) save(id, mtype string, delta models.Counter, value models.Gauge) {
	queryTmp := `insert into "ypt" ("id", "mtype", "delta", "value") values('%s', '%s', %v, %v)`
	query := fmt.Sprintf(queryTmp, id, mtype, delta, value)
	log.Debug().Msg("save query: " + query)
	_, err := s.Db().Exec(query)
	if err != nil {
		log.Info().Err(err).Msg("save item error")
	}
}

func (s *DbData) update(id, mtype string, delta models.Counter, value models.Gauge) {
	query := `update "ypt" set "id"=$1, "mtype"=$2, "delta"=$3, "value"=$4 where "id"=$5`
	log.Debug().Msg("update query: " + query)
	_, err := s.Db().Exec(query, id, mtype, delta, value, id)
	if err != nil {
		log.Info().Err(err).Msg("update item error")
	}
}

func getDbSafelyDelta(link *int64) int64 {
	if link == nil {
		return 0
	}
	return *link
}

func getDbSafelyValue(link *float64) float64 {
	if link == nil {
		return 0.0
	}
	return *link
}
