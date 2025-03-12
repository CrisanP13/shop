package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/crisanp13/shop/src/types"
	"github.com/crisanp13/shop/src/util"
)

func Run(log *log.Logger,
	port string,
	ctx context.Context) error {
	var db *sql.DB
	db, err := createDB()
	_ = db
	if err != nil {
		return fmt.Errorf("failed to create db, %w", err)
	}
	log.Println("connected to db")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealthcheck)
	mux.HandleFunc("POST /user/register", createUserRegisterHandler(log, db))
	mux.HandleFunc("POST /user/login", createUserLoginHandler(log, db))
	log.Println("starting on", port)
	err = http.ListenAndServe(port, mux)
	if err != nil {
		return fmt.Errorf("failed to start sever, %w", err)
	}

	return nil
}

func createDB() (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.User = "root"
	cfg.Passwd = "qwer"
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "shop"
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

func createUserRegisterHandler(log *log.Logger,
	db *sql.DB,
) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		log.Println("received register")
		req, problems, err := util.Decode[types.RegisterReq](r)
		if err != nil {
			if len(problems) > 0 {
				util.Encode(w, r, http.StatusBadRequest,
					types.ErrorResp{Error: problems})
				return
			}
			util.Encode(w, r, http.StatusBadRequest,
				types.ErrorRespFromString("failed to decode json"))
			return
		}

		var count int
		err = db.QueryRow("select count(1) from users where email = ?", req.Email).Scan(&count)
		if err != nil {
			log.Println("failed query,", err)
			util.Encode(w, r, http.StatusInternalServerError,
				types.ErrorRespFromString("internal server error"))
		}
		if count > 0 {
			log.Println("email already in use,", req.Email)
			util.Encode(w, r, http.StatusBadRequest,
				types.ErrorRespFromString("email already in use"))
			return
		}

		pass, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
		if err != nil {
			log.Println("failed to hash password,", err)
			util.Encode(w, r, http.StatusInternalServerError,
				types.ErrorRespFromString("internal server error"))
		}
		res, err := db.Exec("insert into users (name, email, password) values (?, ?, ?)",
			req.Name, req.Email, pass)
		if err != nil {
			log.Printf("failed to create user %+v, %s", req, err)
			util.Encode(w, r, http.StatusInternalServerError,
				types.ErrorRespFromString("internal server error"))
		}
		id, err := res.LastInsertId()
		if err != nil {
			log.Printf("failed to retrieve id of newly created user, %s", err)
			util.Encode(w, r, http.StatusInternalServerError,
				types.ErrorRespFromString("internal server error"))
		}

		util.Encode(w, r, http.StatusCreated,
			types.RegiesterResp{Id: fmt.Sprint(id)})
	}
}

func createUserLoginHandler(log *log.Logger,
	db *sql.DB,
) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		log.Println("received login")
		req, problems, err := util.Decode[types.LoginReq](r)

		if err != nil {
			if len(problems) > 0 {
				util.Encode(w, r, http.StatusBadRequest,
					types.ErrorResp{Error: problems})
				return
			}
			util.Encode(w, r, http.StatusBadRequest,
				types.ErrorResp{Error: "failed to decode json"})
			return
		}

		var password []byte
		var id int
		row := db.QueryRow("select id, password from users where email = ?", req.Email)
		if err = row.Scan(&id, &password); err != nil {
			log.Println("failed login query,", err)
			util.Encode(w, r, http.StatusBadRequest,
				types.ErrorRespFromString("internal server error"))
			return
		}
		err = bcrypt.CompareHashAndPassword(password, []byte(req.Password))
		if err != nil {
			util.Encode(w, r, http.StatusNotFound,
				types.ErrorRespFromString("user not found"))
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256,
			jwt.MapClaims{
				"id":  id,
				"exp": time.Now().Add(time.Hour * 24).Unix(),
			})
		tokenString, err := token.SignedString([]byte("zecret"))
		if err != nil {
			log.Println("failed to generate token,", err)
			util.Encode(w, r, http.StatusInternalServerError,
				types.ErrorRespFromString("internal server error"))
			return
		}

		util.Encode(w, r, http.StatusOK,
			types.LoginResp{Token: "Bearer: " + tokenString})
	}
}

func handleHealthcheck(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusOK)
}
