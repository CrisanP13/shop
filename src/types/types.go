package types

import (
	"net/mail"
	"strings"
)

type ErrorResp struct {
	Error any `json:"error"`
}

func ErrorRespFromString(s string) ErrorResp {
	return ErrorResp{Error: s}
}

var InternalServerError = ErrorResp{Error: "internal server error"}

type RegisterReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r RegisterReq) Validate() map[string]string {
	problems := map[string]string{}

	if strings.TrimSpace(r.Name) == "" {
		problems["name"] = "empty"
	}

	if strings.TrimSpace(r.Email) == "" {
		problems["email"] = "empty"
	} else if _, err := mail.ParseAddress(r.Email); err != nil {
		problems["email"] = "invalid email"
	}

	if strings.TrimSpace(r.Password) == "" {
		problems["password"] = "empty"
	}

	return problems
}

type RegiesterResp struct {
	Id string `json:"id"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginReq) Validate() map[string]string {
	problems := map[string]string{}

	if strings.TrimSpace(r.Email) == "" {
		problems["email"] = "empty"
	}
	if strings.TrimSpace(r.Password) == "" {
		problems["password"] = "empty"
	}

	return problems
}

type LoginResp struct {
	Id    string `json:"id"`
	Token string `json:"token"`
}

type User struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
