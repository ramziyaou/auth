package mysql

import (
	"auth/domain"
	"auth/myerrors"
	"auth/user/repository"
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

type mySQLDBInterface struct {
	db *sql.DB
}

func (m *mySQLDBInterface) GetUser(username string) (*domain.User, error) {
	user := new(domain.User)
	err := m.db.QueryRow("select * from users where username=?", username).Scan(&user.ID, &user.Ts, &user.IIN, &user.Username, &user.Password)
	if err == sql.ErrNoRows {
		return user, myerrors.ErrUserNotFound
	}
	return user, err
}

func (m *mySQLDBInterface) AddUser(IIN, username, password string) error {
	if IIN == "" || username == "" || password == "" {
		return myerrors.ErrInvalidInput
	}
	insForm, err := m.db.Prepare("insert into users (iin, username, password) values(?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = insForm.Exec(IIN, username, password)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return myerrors.ErrDuplicateUser
	}
	return err
}

func (m *mySQLDBInterface) GetUserByIIN(IIN string) (*domain.User, error) {
	user := new(domain.User)
	err := m.db.QueryRow("select * from users where iin=?", IIN).Scan(&user.ID, &user.Ts, &user.IIN, &user.Username, &user.Password)
	if err == sql.ErrNoRows {
		return user, myerrors.ErrUserNotFound
	}
	return user, err
}

func NewMySQLDBInterface() (repository.DBInterface, error) {
	db, err := sql.Open("mysql", os.Getenv("DATA_SOURCE"))
	if err != nil {
		return nil, err
	}
	log.Println("INFO|Success in opening DB")
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(time.Minute * 5)
	db.SetConnMaxIdleTime(time.Minute * 2)
	start := time.Now()
	for db.Ping() != nil {
		if time.Now().After(start.Add(time.Minute * 20)) {
			log.Println("ERROR|Failed to connect after 20 minutes")
			return nil, db.Ping()
		}
	}
	log.Println("INFO|DB Pong", db.Ping() == nil)
	return &mySQLDBInterface{db: db}, nil
}

func (db *mySQLDBInterface) Close() {
	db.db.Close()
}
