package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/go-sql-driver/mysql"

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
	mux.HandleFunc("/user/register", createUserRegisterHandler(log, db))
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
		req, problems, err := util.Decode[types.RegisterReq](r)
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

		var count int
		err = db.QueryRow("select count(1) from users where email = ?", req.Email).Scan(&count)
		if err != nil {
			log.Println("failed failed query,", err)
			util.Encode(w, r, http.StatusInternalServerError,
				types.ErrorRespFromString("internal server error"))
		}
		if count > 0 {
			util.Encode(w, r, http.StatusBadRequest,
				types.ErrorResp{Error: "email already in use"})
			return
		}

		// todo hash pass
		res, err := db.Exec("insert into users (name, email, password) values (?, ?, ?)",
			req.Name, req.Email, req.Password)
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

func handleUserRegister(
	w http.ResponseWriter,
	r *http.Request,
) {
	req, problems, err := util.Decode[types.RegisterReq](r)
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

	log.Default().Printf("%+v", req)

	util.Encode(w, r, http.StatusCreated,
		types.RegiesterResp{Id: "1234"})
}

func handleHealthcheck(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusOK)
}
