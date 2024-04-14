package banner

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/KretovDmitry/avito-tech/internal/config"
	"github.com/KretovDmitry/avito-tech/pkg/log"
	"github.com/redis/go-redis/v9"
)

type Repository interface {
	GetActiveBannerByFeatureTag(ctx context.Context, params GetUserBannerParams) (*Banner, error)
	GetBannersByFeature(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByFeatureWithLimit(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByFeatureWithOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByFeatureWithLimitOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByTag(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByTagWithLimit(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByTagWithOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByTagWithLimitOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByFeatureTag(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByFeatureTagWithLimit(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByFeatureTagWithOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	GetBannersByFeatureTagWithLimitOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error)
	CreateBanner(ctx context.Context, data PostBannerJSONBody) (*PostBannerResponse, error)
	DeleteBannerByID(ctx context.Context, id int) error
	DeleteBannersByID(ctx context.Context, id ...int) error
	UpdateBannerByID(ctx context.Context, id int, data *map[string]interface{}) error
	UpdateIsActiveByID(ctx context.Context, id int, isActive bool) error
	UpdateFeatureByID(ctx context.Context, id int, featureID int) error
	UpdateBannerTagByID(ctx context.Context, id int, tags *[]int) error
}

type repository struct {
	db      *sql.DB
	rdb     *redis.Client
	queries *Queries
	logger  log.Logger
	config  *config.Config
}

func NewRepository(db *sql.DB, rdb *redis.Client, logger log.Logger, config *config.Config) (*repository, error) {
	if db == nil {
		return nil, errors.New("nil dependency: database")
	}
	if config == nil {
		return nil, errors.New("nil dependency: config")
	}

	return &repository{
		db:      db,
		rdb:     rdb,
		queries: New(db),
		logger:  logger,
		config:  config,
	}, nil
}

var _ Repository = (*repository)(nil)

func (r *repository) GetActiveBannerByFeatureTag(ctx context.Context, params GetUserBannerParams) (*Banner, error) {
	key := fmt.Sprintf("feature_id:%d-tag_id:%d", params.FeatureId, params.TagId)

	if params.UseLastRevision != nil && !*params.UseLastRevision {
		data, err := r.rdb.Get(ctx, key).Bytes()
		if err != nil && err != redis.Nil {
			return nil, err
		}

		if err == nil {
			banner := new(Banner)
			if err := json.Unmarshal(data, banner); err != nil {
				return nil, err
			}

			return banner, nil
		}
	}

	banner, err := r.queries.GetActiveBannerByFeatureTag(ctx,
		GetActiveBannerByFeatureTagParams{
			FeatureID: params.FeatureId,
			TagID:     params.TagId,
		})
	if err != nil {
		return nil, err
	}

	buf, _ := json.Marshal(banner)

	err = r.rdb.Set(ctx, key, buf, r.config.CacheExpiration).Err()
	if err != nil {
		return nil, err
	}

	return &banner, nil
}

func (r *repository) GetBannersByFeature(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	banners, err := r.queries.GetBannersByFeature(ctx, *params.FeatureId)
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, b := range banners {
		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByFeatureWithLimit(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	banners, err := r.queries.GetBannersByFeatureWithLimit(ctx,
		GetBannersByFeatureWithLimitParams{
			FeatureID: *params.FeatureId,
			Limit:     *params.Limit,
		})
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, b := range banners {
		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByFeatureWithOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	banners, err := r.queries.GetBannersByFeatureWithOffset(ctx,
		GetBannersByFeatureWithOffsetParams{
			FeatureID: *params.FeatureId,
			Offset:    *params.Offset,
		})
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, b := range banners {
		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByFeatureWithLimitOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	banners, err := r.queries.GetBannersByFeatureWithLimitOffset(ctx,
		GetBannersByFeatureWithLimitOffsetParams{
			FeatureID: *params.FeatureId,
			Limit:     *params.Limit,
			Offset:    *params.Offset,
		})
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, b := range banners {
		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByTag(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	tags, err := r.queries.GetBannersIDsByTag(ctx, *params.TagId)
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, tag := range tags {
		b, err := r.queries.GetBannerByID(ctx, tag.BannerID)
		if err != nil {
			return nil, err
		}

		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByTagWithLimit(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	tags, err := r.queries.GetBannersIDsByTagWithLimit(ctx,
		GetBannersIDsByTagWithLimitParams{
			TagID: *params.TagId,
			Limit: *params.Limit,
		})
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, tag := range tags {
		b, err := r.queries.GetBannerByID(ctx, tag.BannerID)
		if err != nil {
			return nil, err
		}

		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByTagWithOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	tags, err := r.queries.GetBannersIDsByTagWithOffset(ctx,
		GetBannersIDsByTagWithOffsetParams{
			TagID:  *params.TagId,
			Offset: *params.Offset,
		})
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, tag := range tags {
		b, err := r.queries.GetBannerByID(ctx, tag.BannerID)
		if err != nil {
			return nil, err
		}

		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByTagWithLimitOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	tags, err := r.queries.GetBannersIDsByTagWithLimitOffset(ctx,
		GetBannersIDsByTagWithLimitOffsetParams{
			TagID:  *params.TagId,
			Limit:  *params.Limit,
			Offset: *params.Offset,
		})
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, tag := range tags {
		b, err := r.queries.GetBannerByID(ctx, tag.BannerID)
		if err != nil {
			return nil, err
		}

		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByFeatureTag(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	banners, err := r.queries.GetBannersByFeatureTag(ctx,
		GetBannersByFeatureTagParams{
			FeatureID: *params.FeatureId,
			TagID:     *params.TagId,
		})
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, b := range banners {
		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByFeatureTagWithLimit(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	banners, err := r.queries.GetBannersByFeatureTagWithLimit(ctx,
		GetBannersByFeatureTagWithLimitParams{
			FeatureID: *params.FeatureId,
			TagID:     *params.TagId,
			Limit:     *params.Limit,
		})
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, b := range banners {
		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByFeatureTagWithOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	banners, err := r.queries.GetBannersByFeatureTagWithOffset(ctx,
		GetBannersByFeatureTagWithOffsetParams{
			FeatureID: *params.FeatureId,
			TagID:     *params.TagId,
			Offset:    *params.Offset,
		})
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, b := range banners {
		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) GetBannersByFeatureTagWithLimitOffset(ctx context.Context, params GetBannerParams) ([]GetBannerResponse, error) {
	banners, err := r.queries.GetBannersByFeatureTagWithLimitOffset(ctx,
		GetBannersByFeatureTagWithLimitOffsetParams{
			FeatureID: *params.FeatureId,
			TagID:     *params.TagId,
			Limit:     *params.Limit,
			Offset:    *params.Offset,
		})
	if err != nil {
		return nil, err
	}

	response := make([]GetBannerResponse, 0)

	for _, b := range banners {
		t, err := r.queries.GetTagsByBannerID(ctx, b.ID)
		if err != nil {
			return nil, err
		}

		tags := make([]int, 0, 1)
		for _, tag := range t {
			tags = append(tags, tag.TagID)
		}

		response = append(response, GetBannerResponse{
			BannerID:  b.ID,
			FeatureID: b.FeatureID,
			TagIDs:    tags,
			Content:   b,
			IsActive:  b.IsActive,
			CreatedAt: b.CreatedAt,
			UpdatedAt: b.UpdatedAt,
		})
	}

	return response, nil
}

func (r *repository) CreateBanner(ctx context.Context, data PostBannerJSONBody) (*PostBannerResponse, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			r.logger.Error(err)
		}
	}()

	qtx := r.queries.WithTx(tx)

	banner := *data.Content
	fields := []string{"title", "text", "url"}
	vals := make(map[string]string, len(fields))

	for _, field := range fields {
		val, ok := banner[field].(string)
		if !ok {
			return nil, &InvalidTypeError{field}
		}
		vals[field] = val
	}

	id, err := qtx.CreateBanner(ctx, CreateBannerParams{
		FeatureID: *data.FeatureId,
		Title:     vals["title"],
		Text:      vals["text"],
		Url:       vals["url"],
		IsActive:  *data.IsActive,
	})
	if err != nil {
		return nil, err
	}

	for _, tagID := range *data.TagIds {
		_, err = qtx.CreateTag(ctx, CreateTagParams{
			TagID:    tagID,
			BannerID: id,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &PostBannerResponse{BannerID: id}, nil
}

func (r *repository) DeleteBannerByID(ctx context.Context, id int) error {
	_, err := r.queries.DeleteBannerByID(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) DeleteBannersByID(ctx context.Context, ids ...int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			r.logger.Error(err)
		}
	}()

	qtx := r.queries.WithTx(tx)

	for _, id := range ids {
		if _, err := qtx.DeleteBannerByID(ctx, id); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *repository) UpdateBannerByID(ctx context.Context, id int, data *map[string]interface{}) error {
	fields := []string{"title", "text", "url"}
	vals := make(map[string]string, len(fields))

	for _, field := range fields {
		val, ok := (*data)[field].(string)
		if !ok {
			return &InvalidTypeError{field}
		}
		vals[field] = val
	}

	_, err := r.queries.UpdateBannerByID(ctx, UpdateBannerByIDParams{
		Title: vals["title"],
		Text:  vals["text"],
		Url:   vals["url"],
		ID:    id,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) UpdateIsActiveByID(ctx context.Context, id int, isActive bool) error {
	_, err := r.queries.UpdateIsActiveByID(ctx, UpdateIsActiveByIDParams{
		IsActive: isActive,
		ID:       id,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) UpdateFeatureByID(ctx context.Context, id int, featureID int) error {
	_, err := r.queries.UpdateFeatureByID(ctx, UpdateFeatureByIDParams{
		FeatureID: featureID,
		ID:        id,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *repository) UpdateBannerTagByID(ctx context.Context, id int, tags *[]int) error {
	unusedTagID := -1 // mock deletion

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			r.logger.Error(err)
		}
	}()

	qtx := r.queries.WithTx(tx)

	currentTags, err := qtx.GetTagsByBannerID(ctx, id)
	if err != nil {
		return err
	}

	switch {
	case len(currentTags) == len(*tags):
		for i, tag := range *tags {
			_, err := qtx.UpdateBannerTagByID(ctx, UpdateBannerTagByIDParams{
				TagID:    tag,
				BannerID: id,
				ID:       currentTags[i].ID,
			})
			if err != nil {
				return err
			}
		}

	case len(currentTags) < len(*tags):
		for i := 0; i < len(*tags); i++ {
			if i < len(currentTags) {
				_, err := qtx.UpdateBannerTagByID(ctx, UpdateBannerTagByIDParams{
					TagID:    (*tags)[i],
					BannerID: id,
					ID:       currentTags[i].ID,
				})
				if err != nil {
					return err
				}
				continue
			}
			_, err := qtx.CreateTag(ctx, CreateTagParams{
				TagID:    (*tags)[i],
				BannerID: id,
			})
			if err != nil {
				return err
			}
		}

	case len(currentTags) > len(*tags):
		for i := 0; i < len(currentTags); i++ {
			if i < len(*tags) {
				_, err := qtx.UpdateBannerTagByID(ctx, UpdateBannerTagByIDParams{
					TagID:    (*tags)[i],
					BannerID: id,
					ID:       currentTags[i].ID,
				})
				if err != nil {
					return err
				}
				continue
			}
			_, err := qtx.UpdateBannerTagByID(ctx, UpdateBannerTagByIDParams{
				TagID:    unusedTagID,
				BannerID: id,
				ID:       currentTags[i].ID,
			})
			if err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
