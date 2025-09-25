package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sinavarasina/SAS-BOT/internal/db"
)

func HandlerRoutePrivate(dbConn *sqlx.DB, jid, text, username, number string) string {
	text = strings.TrimSpace(text)

	WAUser := db.User{
		JID:       jid,
		Number:    number,
		Username:  username,
		Previlege: "user",
	}

	err := db.SaveUser(dbConn, WAUser)
	if err != nil {
		log.Printf("Error at db.SaveUser for jid: %s, Message : %v", jid, err)
	}

	switch text {
	case "!batal":
		ResetSession(jid)
		return "Sesi Dibatalkan"
	}

	return "Halo, Saya SAS (Sindang Anom Service), \nsaya diutus oleh dewa Reyhan Capri, dan di berkati oleh supremasi tertinggi michael mathew, \nmenerima divine intelect dari arrauf.\nBerikut adalah hal yang dapat saya lakukan :\n\t1. Isi NIK\npilih dengan memilih angka (misal 1) jawab 1 saja."
}

func HandlerRouteGroup(dbConn *sqlx.DB, jid, text, username, number string) string {
	greeter := fmt.Sprintf("Halo, user dengan nickname %s, dengan nomor %s,\nSaya SAS (Sindang Anom Service), harap lakukan sesi private message dengan saia (PMOnly kay?))", username, number)
	return greeter + "\nsaya diutus oleh dewa Reyhan Capri, dan di berkati oleh supremasi tertinggi michael mathew, \nmenerima divine intelect dari arrauf."
}
