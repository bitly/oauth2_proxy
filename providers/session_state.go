package providers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bitly/oauth2_proxy/cookie"
)

type SessionState struct {
	AccessToken  string
	ExpiresOn    time.Time
	RefreshToken string
	Email        string
	User         string
	Groups       []string
}

func (s *SessionState) IsExpired() bool {
	if !s.ExpiresOn.IsZero() && s.ExpiresOn.Before(time.Now()) {
		return true
	}
	return false
}

func (s *SessionState) String() string {
	o := fmt.Sprintf("Session{%s", s.userAndGroups())
	if s.AccessToken != "" {
		o += " token:true"
	}
	if !s.ExpiresOn.IsZero() {
		o += fmt.Sprintf(" expires:%s", s.ExpiresOn)
	}
	if s.RefreshToken != "" {
		o += " refresh_token:true"
	}
	return o + "}"
}

func (s *SessionState) EncodeSessionState(c *cookie.Cipher) (string, error) {
	if c == nil || s.AccessToken == "" {
		return s.userAndGroups(), nil
	}
	return s.EncryptedString(c)
}

func (s *SessionState) userAndGroups() string {
	u := s.User
	if s.Email != "" {
		u = s.Email
	}
	if len(s.Groups) > 0 {
		u += "," + strings.Join(s.Groups, ",")
	}
	return u
}

func decodeUserAndGroups(v string) (user string, email string, groups []string) {
	vs := strings.Split(v, ",")
	if u := vs[0]; strings.Contains(u, "@") {
		email = vs[0]
		user = strings.Split(u, "@")[0]
	} else {
		user = u
	}
	groups = vs[1:]
	return
}

func (s *SessionState) EncryptedString(c *cookie.Cipher) (string, error) {
	var err error
	if c == nil {
		panic("error. missing cipher")
	}
	a := s.AccessToken
	if a != "" {
		a, err = c.Encrypt(a)
		if err != nil {
			return "", err
		}
	}
	r := s.RefreshToken
	if r != "" {
		r, err = c.Encrypt(r)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s|%s|%d|%s", s.userAndGroups(), a, s.ExpiresOn.Unix(), r), nil
}

func DecodeSessionState(v string, c *cookie.Cipher) (s *SessionState, err error) {
	chunks := strings.Split(v, "|")
	s = &SessionState{}

	if len(chunks) == 1 {
		s.User, s.Email, s.Groups = decodeUserAndGroups(chunks[0])
		return
	}

	if len(chunks) != 4 {
		err = fmt.Errorf("invalid number of fields (got %d expected 4)", len(chunks))
		return
	}

	if c != nil && chunks[1] != "" {
		s.AccessToken, err = c.Decrypt(chunks[1])
		if err != nil {
			return nil, err
		}
	}
	if c != nil && chunks[3] != "" {
		s.RefreshToken, err = c.Decrypt(chunks[3])
		if err != nil {
			return nil, err
		}
	}

	s.User, s.Email, s.Groups = decodeUserAndGroups(chunks[0])
	ts, _ := strconv.Atoi(chunks[2])
	s.ExpiresOn = time.Unix(int64(ts), 0)
	return
}
