package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type UserNIK struct {
	JID string `db:"jid"`
	NIK string `db:"nik"`
}

func SaveNIK(dbConn *sqlx.DB, userNik UserNIK) error {
	_, err := dbConn.Exec(`
	INSERT INTO user_nik (jid, nik)
	VALUES ($1, $2)
	ON CONFLICT (jid) DO UPDATE
	SET nik = EXCLUDED.nik`,
		userNik.JID, userNik.NIK)
	return err
}
