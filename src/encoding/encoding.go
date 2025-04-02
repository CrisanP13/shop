package encoding

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/crisanp13/shop/src/types"
)

type Validateable interface {
	Validate() (problems map[string]string)
}

func Encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}
	return nil
}

func EncodeInternalServerError(w http.ResponseWriter) {
	Encode(w, http.StatusInternalServerError,
		types.ErrorResp{Error: "internal server error"})
}

func Decode[T Validateable](r *http.Request) (T, map[string]string, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, nil, fmt.Errorf("failed to decode: %w", err)
	}
	if problems := v.Validate(); len(problems) > 0 {
		return v, problems, fmt.Errorf("invalid %T: %d problems", v, len(problems))
	}
	return v, nil, nil
}
