package banner

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/KretovDmitry/avito-tech/internal/config"
	"github.com/KretovDmitry/avito-tech/internal/user"
	"github.com/KretovDmitry/avito-tech/pkg/log"
)

// Banner service implementation.
type BannerService struct {
	repo       Repository
	logger     log.Logger
	deleteChan chan int
	wg         *sync.WaitGroup
	done       chan struct{}
	config     *config.Config
}

// Make sure we conform to ServerInterface
var _ ServerInterface = (*BannerService)(nil)

// New constructs a new banner service instance,
// ensuring that the dependencies are valid values.
func NewService(repo Repository, logger log.Logger, config *config.Config) (*BannerService, error) {
	if config == nil {
		return nil, errors.New("nil dependency: config")
	}

	service := &BannerService{
		repo:       repo,
		logger:     logger,
		deleteChan: make(chan int, config.BannerBufferLength),
		wg:         &sync.WaitGroup{},
		done:       make(chan struct{}),
		config:     config,
	}

	service.wg.Add(1)
	go func() {
		defer service.wg.Done()
		service.flushDelete()
	}()

	return service, nil
}

// ErrorHandlerFunc wraps sending of an error in the Error format, and
// handling the failure to marshal that.
func ErrorHandlerFunc(w http.ResponseWriter, r *http.Request, err error) {
	bannerError := Error{Error: err.Error()}
	var code int
	switch err.(type) {
	case *RequiredParamError, *RequiredHeaderError,
		*InvalidParamFormatError, *TooManyValuesForParamError,
		*InvalidTypeError:
		code = http.StatusBadRequest
	default:
		code = http.StatusInternalServerError
	}
	// empty body
	if errors.Is(err, io.EOF) {
		code = http.StatusBadRequest
	}
	w.WriteHeader(code)
	if err = json.NewEncoder(w).Encode(bannerError); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *BannerService) Stop() {
	sync.OnceFunc(func() {
		close(s.done)
	})()

	ready := make(chan struct{})
	go func() {
		defer close(ready)
		s.wg.Wait()
	}()

	select {
	case <-time.After(s.config.ShutdownTimeout):
		s.logger.Error("banner service stop: shutdown timeout exceeded")
	case <-ready:
		return
	}
}

func (s *BannerService) flushDelete() {
	ticker := time.NewTicker(10 * time.Second)
	ids := make([]int, 0, s.config.BannerBufferLength)

	for {
		select {
		case id := <-s.deleteChan:
			ids = append(ids, id)

		case <-s.done:
			if len(ids) == 0 {
				return
			}
			_ = s.flush(ids...)
			return

		case <-ticker.C:
			if len(ids) == 0 {
				continue
			}
			if err := s.flush(ids...); err != nil {
				continue
			}
			// reset buffer only when flush succeeded
			ids = ids[:0:s.config.BannerBufferLength]
		}
	}
}

func (s *BannerService) flush(bannerIDs ...int) error {
	if len(bannerIDs) == 0 {
		return nil
	}

	err := s.repo.DeleteBannersByID(context.TODO(), bannerIDs...)
	if err != nil {
		return err
	}

	return nil
}

type GetBannerResponse struct {
	BannerID  int       `json:"banner_id"`
	FeatureID int       `json:"feature_id"`
	TagIDs    []int     `json:"tag_ids"`
	Content   Banner    `json:"content"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Получение всех баннеров c фильтрацией по фиче и/или тегу
// (GET /banner)
func (s *BannerService) GetBanner(w http.ResponseWriter, r *http.Request, params GetBannerParams) {
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if u.Role != "ADMIN" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	response := make([]GetBannerResponse, 0)

	// Everywhere where no results are returned from the database,
	// we return empty slice with status OK.
	// The status not found is not required by the specification.
	switch {
	// nothing is provided, return empty slice
	case params.FeatureId == nil && params.TagId == nil:
		break

	// only feature id is provided
	case params.FeatureId != nil && params.TagId == nil &&
		params.Limit == nil && params.Offset == nil:
		res, err := s.repo.GetBannersByFeature(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// feature id + limit
	case params.FeatureId != nil && params.Limit != nil &&
		params.TagId == nil && params.Offset == nil:
		res, err := s.repo.GetBannersByFeatureWithLimit(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// feature id + offset
	case params.FeatureId != nil && params.Offset != nil &&
		params.TagId == nil && params.Limit == nil:
		res, err := s.repo.GetBannersByFeatureWithOffset(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// feature id + limit + offset
	case params.FeatureId != nil && params.Limit != nil &&
		params.Offset != nil && params.TagId == nil:
		res, err := s.repo.GetBannersByFeatureWithLimitOffset(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// only tag id is provided
	case params.TagId != nil && params.FeatureId == nil &&
		params.Limit == nil && params.Offset == nil:
		res, err := s.repo.GetBannersByTag(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// tag id + limit
	case params.TagId != nil && params.FeatureId == nil &&
		params.Limit != nil && params.Offset == nil:
		res, err := s.repo.GetBannersByTagWithLimit(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// tag id + offset
	case params.TagId != nil && params.FeatureId == nil &&
		params.Limit == nil && params.Offset != nil:
		res, err := s.repo.GetBannersByTagWithOffset(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// tag id + limit + offset
	case params.TagId != nil && params.FeatureId == nil &&
		params.Limit != nil && params.Offset != nil:
		res, err := s.repo.GetBannersByTagWithLimitOffset(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// feature + tag
	case params.FeatureId != nil && params.TagId != nil &&
		params.Limit == nil && params.Offset == nil:
		res, err := s.repo.GetBannersByFeatureTag(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// feature + tag + limit
	case params.FeatureId != nil && params.TagId != nil &&
		params.Limit != nil && params.Offset == nil:
		res, err := s.repo.GetBannersByFeatureTagWithLimit(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// feature + tag + offset
	case params.FeatureId != nil && params.TagId != nil &&
		params.Limit == nil && params.Offset != nil:
		res, err := s.repo.GetBannersByFeatureTagWithOffset(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res

	// feature + tag + limit + offset
	case params.FeatureId != nil && params.TagId != nil &&
		params.Limit != nil && params.Offset != nil:
		res, err := s.repo.GetBannersByFeatureTagWithLimitOffset(r.Context(), params)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			ErrorHandlerFunc(w, r, err)
			return
		}
		response = res
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		ErrorHandlerFunc(w, r, err)
	}
}

type PostBannerResponse struct {
	BannerID int `json:"banner_id"`
}

// Создание нового баннера
// (POST /banner)
func (s *BannerService) PostBanner(w http.ResponseWriter, r *http.Request, params PostBannerParams) {
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if u.Role != "ADMIN" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	data := new(PostBannerJSONBody)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}

	if data.Content == nil || data.FeatureId == nil ||
		data.IsActive == nil || data.TagIds == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := s.repo.CreateBanner(r.Context(), *data)
	if err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		ErrorHandlerFunc(w, r, err)
	}
}

// Удаление баннеров по id
// (DELETE /banner)
func (s *BannerService) DeleteBanner(w http.ResponseWriter, r *http.Request, params DeleteBannerParams) {
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if u.Role != "ADMIN" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ids := make(DeleteBannerJSONBody, 0)
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}

	for _, id := range ids {
		s.deleteChan <- id
	}

	w.WriteHeader(http.StatusAccepted)
}

// Удаление баннера по идентификатору
// (DELETE /banner/{id})
func (s *BannerService) DeleteBannerId(w http.ResponseWriter, r *http.Request, id int, params DeleteBannerIdParams) {
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if u.Role != "ADMIN" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err := s.repo.DeleteBannerByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		ErrorHandlerFunc(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Обновление содержимого баннера
// (PATCH /banner/{id})
func (s *BannerService) PatchBannerId(w http.ResponseWriter, r *http.Request, id int, params PatchBannerIdParams) {
	u, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if u.Role != "ADMIN" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	data := new(PatchBannerIdJSONBody)
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		ErrorHandlerFunc(w, r, err)
		return
	}

	if data.Content != nil {
		if err := s.repo.UpdateBannerByID(r.Context(), id, data.Content); err != nil {
			ErrorHandlerFunc(w, r, err)
			return
		}
	}

	if data.IsActive != nil {
		if err := s.repo.UpdateIsActiveByID(r.Context(), id, *data.IsActive); err != nil {
			ErrorHandlerFunc(w, r, err)
			return
		}
	}

	if data.FeatureId != nil {
		if err := s.repo.UpdateFeatureByID(r.Context(), id, *data.FeatureId); err != nil {
			ErrorHandlerFunc(w, r, err)
			return
		}
	}

	if data.TagIds != nil {
		if err := s.repo.UpdateBannerTagByID(r.Context(), id, data.TagIds); err != nil {
			ErrorHandlerFunc(w, r, err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Получение баннера для пользователя
// (GET /user_banner)
func (s *BannerService) GetUserBanner(w http.ResponseWriter, r *http.Request, params GetUserBannerParams) {
	_, found := user.FromContext(r.Context())
	if !found {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	banner, err := s.repo.GetActiveBannerByFeatureTag(r.Context(), params)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		ErrorHandlerFunc(w, r, err)
		return
	}

	if err = json.NewEncoder(w).Encode(banner); err != nil {
		ErrorHandlerFunc(w, r, err)
	}
}
