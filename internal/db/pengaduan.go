package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Pengaduan struct {
	UserJID   string `db:"user_jid"`
	Deskripsi string `db:"deskripsi"`
	PictPath  string `db:"pict_path"`
}

func SavePengaduan(dbConn *sqlx.DB, aduan Pengaduan) (int, error) {
	var newID int
	err := dbConn.QueryRowx(`
	INSERT INTO pengaduan (user_jid, deskripsi, pict_path)
	VALUES ($1, $2, $3)
	RETURNING id`,
		aduan.UserJID, aduan.Deskripsi, aduan.PictPath).Scan(&newID)
	return newID, err
}
