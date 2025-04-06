package stores

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/crisanp13/shop/src/types"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func CreateDb(getEnv func(string) string) (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.User = getEnv("SHOP_DB_USER")
	cfg.Passwd = getEnv("SHOP_DB_PASS")
	cfg.Net = getEnv("SHOP_DB_NET")
	cfg.Addr = getEnv("SHOP_DB_ADDR")
	cfg.DBName = getEnv("SHOP_DB_NAME")
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open db, %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping db, %w", err)
	}
	return db, nil
}

type UserStore interface {
	EmailExists(string) (bool, error)
	Create(*types.User, string) (string, error)
	GetIdAndPassByEmail(string) (string, []byte, error)
	GetById(string) (*types.User, error)
}

type MySqlUserStore struct {
	db *sql.DB
}

func NewMySqlUserStore(db *sql.DB) *MySqlUserStore {
	return &MySqlUserStore{db: db}
}

func (s *MySqlUserStore) EmailExists(e string) (bool, error) {
	var count int
	err := s.db.QueryRow("select count(1) from users where email = ?", e).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed query, %w", err)
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

func (s MySqlUserStore) Create(u *types.User, p string) (string, error) {
	pass, err := bcrypt.GenerateFromPassword([]byte(p), 14)
	if err != nil {
		return "", fmt.Errorf("failed to hash password, %w", err)
	}
	res, err := s.db.Exec("insert into users (name, email, password) values (?, ?, ?)",
		u.Name, u.Email, pass)
	if err != nil {
		return "", fmt.Errorf("failed to create user %+v, %s", u, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to retrieve id of newly created user, %s", err)
	}
	return strconv.FormatInt(id, 10), nil
}

func (s MySqlUserStore) GetIdAndPassByEmail(e string) (string, []byte, error) {
	row := s.db.QueryRow("select id, password from users where email = ?", e)
	var id string
	var pass []byte
	err := row.Scan(&id, &pass)
	if err == sql.ErrNoRows {
		return "", nil, nil
	}
	if err != nil {
		return "", nil, fmt.Errorf("failed to retrieve user by email, %w", err)
	}
	return id, pass, nil
}

func (s MySqlUserStore) GetById(id string) (*types.User, error) {
	row := s.db.QueryRow("select id, name, email from users where id = ?", id)
	var user types.User
	if err := row.Scan(&user.Id, &user.Name, &user.Email); err != nil {
		return nil, fmt.Errorf("failed to retrieve user by id, %w", err)
	}
	return &user, nil
}
