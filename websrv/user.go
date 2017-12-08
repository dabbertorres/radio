package websrv

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	newUserCmd     = "insert into users (name, password) values ($1, $2)"
	loginUserQuery = "select password from users where name = $1"
)

type User struct {
	Name string `json:"name"`
	Email string `json:"email"`
}

func NewUser(db *sql.DB, username, email, password string) (User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}

	_, err = db.Exec(newUserCmd, username, hashed)
	if err != nil {
		// TODO check if error is due to 'username' already in database
		return User{}, err
	}

	return User{
		Name: username,
		Email: email,
	}, nil
}

func Login(db *sql.DB, username, password, ipAddr, userAgent string, duration time.Duration) (SessionId, error) {
	row := db.QueryRow(loginUserQuery, username)

	hashed := ""
	if err := row.Scan(&hashed); err != nil {
		// TODO error filtering/handling
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return "", err
	}

	return NewSession(db, username, ipAddr, userAgent, duration)
}

func Logout(db *sql.DB, id SessionId) error {
	// TODO shutdown user's station

	return EndSession(db, id)
}
