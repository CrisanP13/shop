package api

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/crisanp13/shop/src/encoding"
	"github.com/crisanp13/shop/src/stores"
	"github.com/crisanp13/shop/src/types"
)

type IdContextKey string

var idContextKey = IdContextKey("id")
var privateKey = []byte("zecret")

func Run(log *log.Logger,
	port string,
	ctx context.Context) error {
	var db *sql.DB
	db, err := creteDb()
	if err != nil {
		return fmt.Errorf("failed to create db, %w", err)
	}
	us := stores.NewMySqlUserStore(db)
	log.Println("connected to db")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealthcheck)
	mux.HandleFunc("POST /user/register", createUserRegisterHandler(log, us))
	mux.HandleFunc("POST /user/login", createUserLoginHandler(log, us))
	mux.Handle("GET /user/details/{id}",
		authorizationMiddleware(userDetailsHandler(log, us)))
	log.Println("starting on", port)
	err = http.ListenAndServe(port, mux)
	if err != nil {
		return fmt.Errorf("failed to start sever, %w", err)
	}

	return nil
}

func creteDb() (*sql.DB, error) {
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

func handleHealthcheck(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusOK)
}

func createUserRegisterHandler(log *log.Logger,
	us stores.UserStore,
) func(http.ResponseWriter, *http.Request) {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		log.Println("received register")
		req, problems, err := encoding.Decode[types.RegisterReq](r)
		if err != nil {
			if len(problems) > 0 {
				encoding.Encode(w, http.StatusBadRequest,
					types.ErrorResp{Error: problems})
				return
			}
			encoding.Encode(w, http.StatusBadRequest,
				types.ErrorRespFromString("failed to decode json"))
			return
		}

		exists, err := us.EmailExists(req.Email)
		if err != nil {
			log.Println("failed email check,", err)
			encoding.Encode(w, http.StatusInternalServerError,
				types.InternalServerError)
			return
		}
		if exists {
			log.Println("email already in use,", req.Email)
			encoding.Encode(w, http.StatusBadRequest,
				types.ErrorRespFromString("email already in use"))
			return
		}

		id, err := us.Create(&types.User{
			Name:  req.Name,
			Email: req.Email},
			req.Password)
		if err != nil {
			log.Printf("failed to create user %+v, %s", req, err)
			encoding.Encode(w, http.StatusInternalServerError,
				types.ErrorRespFromString("internal server error"))
		}

		encoding.Encode(w, http.StatusCreated,
			types.RegiesterResp{Id: id})
	}
}

func createUserLoginHandler(log *log.Logger,
	us stores.UserStore,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("received login")
		req, problems, err := encoding.Decode[types.LoginReq](r)

		if err != nil {
			if len(problems) > 0 {
				encoding.Encode(w, http.StatusBadRequest,
					types.ErrorResp{Error: problems})
				return
			}
			encoding.Encode(w, http.StatusBadRequest,
				types.ErrorResp{Error: "failed to decode json"})
			return
		}
		id, pass, err := us.GetIdAndPassByEmail(req.Email)
		if err != nil {
			log.Println("failed login query,", err)
			encoding.EncodeInternalServerError(w)
			return
		}
		if err != nil {
			encoding.Encode(w, http.StatusNotFound,
				types.ErrorRespFromString("user or password not found"))
			return
		}
		err = bcrypt.CompareHashAndPassword(pass, []byte(req.Password))
		if err != nil {
			encoding.Encode(w, http.StatusNotFound,
				types.ErrorRespFromString("user or password not found"))
			return
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256,
			jwt.MapClaims{
				"id":  id,
				"exp": time.Now().Add(time.Hour * 24).Unix(),
			})
		tokenString, err := token.SignedString(privateKey)
		if err != nil {
			log.Println("failed to generate token,", err)
			encoding.EncodeInternalServerError(w)
			return
		}
		encoding.Encode(w, http.StatusOK,
			types.LoginResp{Token: "Bearer: " + tokenString,
				Id: id})
	}
}

func authorizationMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		tokenString = strings.TrimPrefix(tokenString, "Bearer: ")
		tokenString = strings.TrimSpace(tokenString)
		if tokenString == "" {
			encoding.Encode(w, http.StatusUnauthorized,
				types.ErrorRespFromString("unauthorized"))
			return
		}
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return privateKey, nil
		})
		if err != nil {
			log.Println("failed to parse token,", tokenString, err)
			encoding.Encode(w, http.StatusUnauthorized,
				types.ErrorRespFromString("unauthorized"))
			return
		}
		if !token.Valid {
			log.Println("invalid token", tokenString)
			encoding.Encode(w, http.StatusUnauthorized,
				types.ErrorRespFromString("unauthorized"))
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		ctx := context.WithValue(r.Context(), idContextKey, claims["id"])
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userDetailsHandler(log *log.Logger,
	us stores.UserStore,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("received details")
		id := r.PathValue("id")
		idFromClaims := r.Context().Value(idContextKey)
		log.Println(id)
		log.Println(idFromClaims)
		if id != idFromClaims {
			encoding.Encode(w, http.StatusUnauthorized,
				types.ErrorRespFromString("unauthorized"))
			return
		}
		user, err := us.GetById(id)
		if err != nil {
			encoding.EncodeInternalServerError(w)
			return
		}
		if user == nil {
			encoding.Encode(w, http.StatusNotFound,
				types.ErrorRespFromString("user not found"))
		}
		encoding.Encode(w, http.StatusOK, &user)
	})
}
