-- name: GetUserByID :one
SELECT
    *
FROM
    users
WHERE
    id = $1;

-- name: GetActiveBannerByFeatureTag :one
SELECT
    b.id,
    b.feature_id,
    b.title,
    b.text,
    b.url,
    b.is_active,
    b.is_deleted,
    b.created_at,
    b.updated_at
FROM
    banners b
    JOIN tags t ON t.banner_id = b.id
WHERE
    b.feature_id = $1
    AND b.is_active = TRUE
    AND b.is_deleted = FALSE
    AND t.tag_id = $2;

-- name: GetBannerByID :one
SELECT
    *
FROM
    banners
WHERE
    id = $1;

-- name: GetBannersByFeature :many
SELECT
    *
FROM
    banners
WHERE
    feature_id = $1;

-- name: GetTagsByBannerID :many
SELECT
    *
FROM
    tags
WHERE
    banner_id = $1;

-- name: GetBannersIDsByTag :many
SELECT
    *
FROM
    tags
WHERE
    tag_id = $1;

-- name: GetBannersByFeatureWithLimit :many
SELECT
    *
FROM
    banners
WHERE
    feature_id = $1
ORDER BY
    id
LIMIT $2;

-- name: GetBannersByFeatureWithOffset :many
SELECT
    *
FROM
    banners
WHERE
    feature_id = $1
ORDER BY
    id OFFSET $2;

-- name: GetBannersByFeatureWithLimitOffset :many
SELECT
    *
FROM
    banners
WHERE
    feature_id = $1
ORDER BY
    id
LIMIT $2 OFFSET $3;

-- name: GetBannersIDsByTagWithLimit :many
SELECT
    *
FROM
    tags
WHERE
    tag_id = $1
ORDER BY
    banner_id
LIMIT $2;

-- name: GetBannersIDsByTagWithOffset :many
SELECT
    *
FROM
    tags
WHERE
    tag_id = $1
ORDER BY
    banner_id OFFSET $2;

-- name: GetBannersIDsByTagWithLimitOffset :many
SELECT
    *
FROM
    tags
WHERE
    tag_id = $1
ORDER BY
    banner_id
LIMIT $2 OFFSET $3;

-- name: GetBannersByFeatureTag :many
SELECT
    b.id,
    b.feature_id,
    b.title,
    b.text,
    b.url,
    b.is_active,
    b.is_deleted,
    b.created_at,
    b.updated_at
FROM
    banners b
    JOIN tags t ON t.banner_id = b.id
WHERE
    b.feature_id = $1
    AND t.tag_id = $2;

-- name: GetBannersByFeatureTagWithLimit :many
SELECT
    b.id,
    b.feature_id,
    b.title,
    b.text,
    b.url,
    b.is_active,
    b.is_deleted,
    b.created_at,
    b.updated_at
FROM
    banners b
    JOIN tags t ON t.banner_id = b.id
WHERE
    b.feature_id = $1
    AND t.tag_id = $2
ORDER BY
    b.id
LIMIT $3;

-- name: GetBannersByFeatureTagWithOffset :many
SELECT
    b.id,
    b.feature_id,
    b.title,
    b.text,
    b.url,
    b.is_active,
    b.is_deleted,
    b.created_at,
    b.updated_at
FROM
    banners b
    JOIN tags t ON t.banner_id = b.id
WHERE
    b.feature_id = $1
    AND t.tag_id = $2
ORDER BY
    b.id OFFSET $3;

-- name: GetBannersByFeatureTagWithLimitOffset :many
SELECT
    b.id,
    b.feature_id,
    b.title,
    b.text,
    b.url,
    b.is_active,
    b.is_deleted,
    b.created_at,
    b.updated_at
FROM
    banners b
    JOIN tags t ON t.banner_id = b.id
WHERE
    b.feature_id = $1
    AND t.tag_id = $2
ORDER BY
    b.id
LIMIT $3 OFFSET $4;

-- name: CreateBanner :one
INSERT INTO banners (feature_id, title, text, url, is_active)
    VALUES ($1, $2, $3, $4, $5)
RETURNING
    id;

-- name: CreateTag :one
INSERT INTO tags (tag_id, banner_id)
    VALUES ($1, $2)
RETURNING
    tag_id;

-- name: DeleteBannerByID :one
UPDATE
    banners
SET
    is_deleted = TRUE
WHERE
    id = $1
RETURNING
    id;

-- name: UpdateBannerTagByID :one
UPDATE
    tags
SET
    tag_id = $1
WHERE
    banner_id = $2
    and id = $3
RETURNING
    id;

-- name: UpdateFeatureByID :one
UPDATE
    banners
SET
    feature_id = $1
WHERE
    id = $2
RETURNING
    id;

-- name: UpdateBannerByID :one
UPDATE
    banners
SET
    title = $1,
    text = $2,
    url = $3
WHERE
    id = $4
RETURNING
    id;

-- name: UpdateIsActiveByID :one
UPDATE
    banners
SET
    is_active = $1
WHERE
    id = $2
RETURNING
    id;

