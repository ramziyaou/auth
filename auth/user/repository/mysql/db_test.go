package mysql

import (
	"auth/domain"
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var getTestTable = []struct {
	name       string
	closeDb    bool
	ErrMessage string
}{
	{"Non-existent user", false, "user not found"},
	{"Closed DB", true, "sql: database is closed"},
}

var u = &domain.User{
	ID:       1,
	Ts:       "2021-12-31 19:36:36",
	IIN:      "910815450350",
	Username: "user",
	Password: "password",
}

var empty_u = &domain.User{}

func NewMock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return db, mock
}

// func TestFindByID(t *testing.T) {
// 	db, mock := NewMock()
// 	defer db.Close()
// 	repo := &mySQLDBInterface{db}

// 	query := "SELECT id, name, email, phone FROM users WHERE id = \\?"

// 	rows := sqlmock.NewRows([]string{"id", "name", "email", "phone"}).
// 		AddRow(u.ID, u.Name, u.Email, u.Phone)

// 	mock.ExpectQuery(query).WithArgs(u.ID).WillReturnRows(rows)

// 	user, err := repo.FindByID(u.ID)
// 	assert.NotNil(t, user)
// 	assert.NoError(t, err)
// }

func TestGetUser(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "select * from users where username=?"

	rows := sqlmock.NewRows([]string{"id", "ts", "iin", "username", "password"}).
		AddRow(u.ID, u.Ts, u.IIN, u.Username, u.Password)

	mock.ExpectQuery(query).WithArgs(u.Username).WillReturnRows(rows)
	user, err := repo.GetUser(u.Username)
	assert.NotNil(t, user)
	assert.NoError(t, err)
}

func TestGetUserError(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "select * from users where username=?"
	rows := sqlmock.NewRows([]string{"id", "ts", "iin", "username", "password"})
	for _, tt := range getTestTable {
		if tt.closeDb {
			db.Close()
		}
		fmt.Println("Running GetUser(username):", tt.name, "******************************************************************************************************")
		mock.ExpectQuery(query).WithArgs(u.Username).WillReturnRows(rows)
		user, err := repo.GetUser(u.Username)
		assert.Empty(t, user)
		assert.EqualError(t, err, tt.ErrMessage)
	}
}

var addTestTable = []struct {
	name       string
	closeDb    bool
	ErrMessage string
}{
	{"Empty user", false, "invalid input"},
	{"Closed DB", true, "sql: database is closed"},
}

func TestAddUser(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "insert into users (iin, username, password) values(?, ?, ?)"

	prep := mock.ExpectPrepare(query)
	prep.ExpectExec().WithArgs(u.IIN, u.Username, u.Password).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.AddUser(u.IIN, u.Username, u.Password)
	assert.NoError(t, err)
}

func TestAddUserError(t *testing.T) {
	user := &domain.User{}
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}
	user = empty_u
	query := "insert into users (iin, username, password) values(?, ?, ?)"
	for _, tt := range addTestTable {
		fmt.Println("Running AddUser:", tt.name, "******************************************************************************************************")
		if tt.closeDb {
			db.Close()
			user = u
		}
		prep := mock.ExpectPrepare(query)
		prep.ExpectExec().WithArgs(user.IIN, user.Username, user.Password).WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.AddUser(user.IIN, user.Username, user.Password)
		assert.EqualError(t, err, tt.ErrMessage)
	}

}

func TestGetUserByIIN(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "select * from users where iin=?"

	rows := sqlmock.NewRows([]string{"id", "ts", "iin", "username", "password"}).
		AddRow(u.ID, u.Ts, u.IIN, u.Username, u.Password)

	mock.ExpectQuery(query).WithArgs(u.IIN).WillReturnRows(rows)
	user, err := repo.GetUserByIIN(u.IIN)
	assert.NotNil(t, user)
	assert.NoError(t, err)
}

func TestGetUserByIINError(t *testing.T) {
	db, mock := NewMock()
	defer db.Close()
	repo := &mySQLDBInterface{db}

	query := "select * from users where iin=?"

	rows := sqlmock.NewRows([]string{"id", "ts", "iin", "username", "password"})
	for _, tt := range getTestTable {
		fmt.Println("Running GetUserByIIN:", tt.name, "******************************************************************************************************")
		if tt.closeDb {
			db.Close()
		}
		mock.ExpectQuery(query).WithArgs(empty_u.IIN).WillReturnRows(rows)
		user, err := repo.GetUserByIIN(empty_u.IIN)
		assert.Empty(t, user)
		assert.EqualError(t, err, tt.ErrMessage)

	}
	// // ErrNotFound
	// mock.ExpectQuery(query).WithArgs(empty_u.IIN).WillReturnRows(rows)
	// user, err := repo.GetUserByIIN(empty_u.IIN)
	// assert.Empty(t, user)
	// assert.EqualError(t, err, "User not found")

	// // Closed DB
	// db.Close()
	// mock.ExpectQuery(query).WithArgs(empty_u.IIN).WillReturnRows(rows)
	// user, err = repo.GetUserByIIN(empty_u.IIN)
	// assert.Empty(t, user)
	// assert.Error(t, err, err.Error())
}
