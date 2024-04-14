package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/KretovDmitry/avito-tech/internal/config"
	"github.com/KretovDmitry/avito-tech/internal/jwt"
	"github.com/KretovDmitry/avito-tech/internal/user"
	"github.com/KretovDmitry/avito-tech/pkg/log"
)

type authService struct {
	repo   Repository
	logger log.Logger
	config *config.Config
}

func NewService(repo Repository, logger log.Logger, config *config.Config) (*authService, error) {
	if config == nil {
		return nil, errors.New("nil dependency: config")
	}
	return &authService{repo: repo, logger: logger, config: config}, nil
}

type JSONError struct {
	Err string `json:"error"`
}

func (a *authService) Middleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		headers := r.Header
		tokenVals, found := headers[http.CanonicalHeaderKey("token")]
		if !found {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if len(tokenVals) > 1 {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(JSONError{"too many values in toker header"})
			return
		}

		userID, err := jwt.GetUserID(tokenVals[0], a.config.JWTSigningKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(JSONError{fmt.Sprintf("invalid token: %v", err)})
			return
		}

		u, err := a.repo.GetUserByID(r.Context(), userID)
		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(JSONError{"no such user"})
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r = r.WithContext(user.NewContext(r.Context(), u))

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}
