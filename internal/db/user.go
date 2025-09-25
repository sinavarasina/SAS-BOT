package db

import "github.com/jmoiron/sqlx"

type User struct {
	JID       string `db:"jid"`
	Number    string `db:"number"`
	Username  string `db:"username"`
	Previlege string `db:"previlege"`
}

func SaveUser(dbConn *sqlx.DB, user User) error {
	_, err := dbConn.Exec(`
	INSERT INTO users (jid, number, username, previlege)
	VALUES (?, ?, ?, ?)
	ON CONFLICT(jid) DO UPDATE SET
		number = excluded.number,
		username = excluded.username,
		previlege = excluded.previlege`,
		user.JID, user.Number, user.Username, user.Previlege)
	return err
}

func GetUser(dbConn *sqlx.DB, jid string) (*User, error) {
	var user User
	err := dbConn.Get(&user, "SELECT * FROM users WHERE jid = ?", jid)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
