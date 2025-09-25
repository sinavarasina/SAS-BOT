package bot

type Session struct {
	Step   string
	Buffer map[string]string
}

var sessions = map[string]*Session{}

func GetSession(jid string) *Session {
	s, ok := sessions[jid]
	if !ok {
		s = &Session{Buffer: make(map[string]string)}
		sessions[jid] = s
	}
	return s
}

func ResetSession(jid string) {
	delete(sessions, jid)
}
