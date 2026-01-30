package models

import (
	"database/sql/driver"

	"github.com/lib/pq"
)

// UUIDArray is a custom type for handling PostgreSQL UUID arrays
type UUIDArray []string

// Scan implements the sql.Scanner interface
func (a *UUIDArray) Scan(src interface{}) error {
	return (*pq.StringArray)(a).Scan(src)
}

// Value implements the driver.Valuer interface
func (a UUIDArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	// Use pq.Array to properly serialize the array
	return pq.Array([]string(a)).Value()
}
