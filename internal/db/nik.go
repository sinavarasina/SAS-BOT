package db

import "github.com/jmoiron/sqlx"

type UserNIK struct {
	JID string `db:"jid"`
	NIK string `db:"nik"`
}

func SaveNIK(dbConn *sqlx.DB, userNik UserNIK) error {
	_, err := dbConn.Exec(`
	INSERT INTO user_nik (jid, nik)
	VALUES (?, ?)
	ON CONFLICT(jid) DO UPDATE SET
		nik = excluded.nik`,
		userNik.JID, userNik.NIK)
	return err
}

// incomplete (let my friend working on it)
