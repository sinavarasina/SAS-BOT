package db

import (
	"database/sql"
	"time"
)

func ToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func ToNullInt64(i int64) sql.NullInt64 {
	// Kita asumsikan 0 adalah "tidak valid" atau "tidak diisi"
	return sql.NullInt64{Int64: i, Valid: i != 0}
}

func ToNullTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: !t.IsZero()}
}
