package stores

import "github.com/crisanp13/shop/src/types"

type UserStore interface {
	GetByEmail(string) types.User
}
