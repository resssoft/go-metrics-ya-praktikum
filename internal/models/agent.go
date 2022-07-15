package models

import "errors"

type Gauge float64
type Counter int64

var ErrNotFound = errors.New("not found")
