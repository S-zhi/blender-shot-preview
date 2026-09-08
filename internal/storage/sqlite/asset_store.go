package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/S-zhi/blender-shot-preview/internal/service"
	api "github.com/S-zhi/blender-shot-preview/kitex_gen/handler/v0_1"
)

type AssetStore struct{ db *sql.DB }

func NewAssetStore(db *sql.DB) *AssetStore { return &AssetStore{db: db} }

func (s *AssetStore) List(ctx context.Context, userID string, assetType *api.AssetType, keyword string, pageNum, pageSize int) ([]service.AssetRecord, int64, error) {
	where := []string{"(user_id = ? OR user_id = 'default_user_001')"}
	args := []any{userID}
	if assetType != nil {
		where = append(where, "asset_type = ?")
		args = append(args, int32(*assetType))
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		where = append(where, "(lower(name) LIKE ? OR lower(description) LIKE ? OR lower(tags_json) LIKE ?)")
		q := "%" + strings.ToLower(keyword) + "%"
		args = append(args, q, q, q)
	}
	clause := strings.Join(where, " AND ")
	var total int64
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM assets WHERE "+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}
	args = append(args, pageSize, (pageNum-1)*pageSize)
	rows, err := s.db.QueryContext(ctx, "SELECT asset_id,user_id,name,asset_type,file_format,file_size_bytes,storage_uri,thumbnail_uri,status,tags_json,created_at,updated_at,description FROM assets WHERE "+clause+" ORDER BY created_at DESC LIMIT ? OFFSET ?", args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var records []service.AssetRecord
	for rows.Next() {
		record, err := scanAsset(rows)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}
	return records, total, rows.Err()
}

func (s *AssetStore) Get(ctx context.Context, userID, assetID string) (service.AssetRecord, bool, error) {
	row := s.db.QueryRowContext(ctx, `SELECT asset_id,user_id,name,asset_type,file_format,file_size_bytes,storage_uri,thumbnail_uri,status,tags_json,created_at,updated_at,description FROM assets WHERE asset_id=? AND (user_id=? OR user_id='default_user_001')`, assetID, userID)
	record, err := scanAsset(row)
	if err == sql.ErrNoRows {
		return service.AssetRecord{}, false, nil
	}
	if err != nil {
		return service.AssetRecord{}, false, err
	}
	return record, true, nil
}

func (s *AssetStore) Create(ctx context.Context, record service.AssetRecord) error {
	tags, _ := json.Marshal(record.Tags)
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO assets(asset_id,user_id,name,asset_type,file_format,file_size_bytes,storage_uri,thumbnail_uri,status,tags_json,created_at,updated_at,description) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, record.AssetID, record.UserID, record.Name, record.AssetType, record.FileFormat, record.FileSizeBytes, record.StorageURI, record.ThumbnailURI, record.Status, string(tags), record.CreatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"), record.UpdatedAt.UTC().Format("2006-01-02T15:04:05.999999999Z07:00"), record.Description)
	return err
}

func (s *AssetStore) Delete(ctx context.Context, userID, assetID string) (bool, error) {
	result, err := s.db.ExecContext(ctx, `DELETE FROM assets WHERE asset_id=? AND (user_id=? OR user_id='default_user_001')`, assetID, userID)
	if err != nil {
		return false, err
	}
	n, _ := result.RowsAffected()
	return n > 0, nil
}

func (s *AssetStore) GetStats(ctx context.Context, userID string) (total, models, presets, totalBytes int64, err error) {
	where := "(user_id=? OR user_id='default_user_001')"
	if err = s.db.QueryRowContext(ctx, "SELECT COUNT(*), COALESCE(SUM(file_size_bytes),0) FROM assets WHERE "+where, userID).Scan(&total, &totalBytes); err != nil {
		return
	}
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM assets WHERE "+where+" AND asset_type=?", userID, api.AssetType_MODEL_3D).Scan(&models)
	if err != nil {
		return
	}
	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM assets WHERE "+where+" AND asset_type=?", userID, api.AssetType_SHOT_PRESET).Scan(&presets)
	return
}

// SeedDefaults keeps the development asset catalogue available after moving
// from the old in-memory store. It is idempotent and never overwrites user data.
func (s *AssetStore) SeedDefaults(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM assets`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	seed := service.NewInMemoryAssetStore()
	records, _, err := seed.List(ctx, "default_user_001", nil, "", 1, 100)
	if err != nil {
		return err
	}
	for _, record := range records {
		if err := s.Create(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scanAsset(row scanner) (service.AssetRecord, error) {
	var r service.AssetRecord
	var tags, created, updated string
	var assetType, status int32
	if err := row.Scan(&r.AssetID, &r.UserID, &r.Name, &assetType, &r.FileFormat, &r.FileSizeBytes, &r.StorageURI, &r.ThumbnailURI, &status, &tags, &created, &updated, &r.Description); err != nil {
		return r, err
	}
	r.AssetType = api.AssetType(assetType)
	r.Status = api.AssetStatus(status)
	_ = json.Unmarshal([]byte(tags), &r.Tags)
	var err error
	r.CreatedAt, err = parseTime(created)
	if err != nil {
		return r, err
	}
	r.UpdatedAt, err = parseTime(updated)
	return r, err
}
func parseTime(value string) (t time.Time, err error) {
	t, err = time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return t, fmt.Errorf("parse sqlite time: %w", err)
	}
	return
}
