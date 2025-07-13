package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/aberyotaro/drink-tracker/models"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
)

type DrinkService struct {
	db *sql.DB
}

func NewDrinkService(db *sql.DB) *DrinkService {
	return &DrinkService{db: db}
}

func (ds *DrinkService) RecordDrink(ctx context.Context, userID int64, drinkType string, amountMl int64, alcoholPercentage float64) (*models.DrinkRecord, error) {
	now := time.Now().UTC()

	// BobのInsertクエリビルダーを使用
	setter := &models.DrinkRecordSetter{
		UserID:            &[]int32{int32(userID)}[0],
		DrinkType:         &drinkType,
		AmountML:          &[]int32{int32(amountMl)}[0],
		AlcoholPercentage: &[]float32{float32(alcoholPercentage)}[0],
		RecordedAt:        &sql.Null[time.Time]{V: now, Valid: true},
		CreatedAt:         &sql.Null[time.Time]{V: now, Valid: true},
	}

	insertQuery := models.DrinkRecords.Insert(setter)
	sqlQuery, args, err := insertQuery.Build(ctx)
	if err != nil {
		return nil, err
	}

	result, err := ds.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	record := &models.DrinkRecord{
		ID:                int32(id),
		UserID:            int32(userID),
		DrinkType:         drinkType,
		AmountML:          int32(amountMl),
		AlcoholPercentage: float32(alcoholPercentage),
		RecordedAt:        sql.Null[time.Time]{V: now, Valid: true},
		CreatedAt:         sql.Null[time.Time]{V: now, Valid: true},
	}

	return record, nil
}

func (ds *DrinkService) GetTodayDrinks(ctx context.Context, userID int64) ([]*models.DrinkRecord, error) {
	// Use datetime range instead of string parsing
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// BobのRAW SQL機能を使用して日付範囲クエリを構築
	query := models.DrinkRecords.Query(
		models.SelectWhere.DrinkRecords.UserID.EQ(int32(userID)),
		sm.Where(sqlite.Raw("recorded_at >= ? AND recorded_at < ?", startOfDay.Format("2006-01-02 15:04:05.999999999 -0700 MST"), endOfDay.Format("2006-01-02 15:04:05.999999999 -0700 MST"))),
		sm.OrderBy(models.DrinkRecordColumns.RecordedAt).Desc(),
	)

	sqlQuery, args, err := query.Build(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := ds.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			// ログに記録するか、適切にハンドリング
		}
	}()

	var records []*models.DrinkRecord
	for rows.Next() {
		var record models.DrinkRecord
		if err := rows.Scan(
			&record.ID, &record.UserID, &record.DrinkType,
			&record.AmountML, &record.AlcoholPercentage,
			&record.RecordedAt, &record.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, &record)
	}

	return records, rows.Err()
}

func (ds *DrinkService) GetTodayTotalAlcohol(ctx context.Context, userID int64) (float64, int64, error) {
	// Use datetime range instead of string parsing
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// BobのRAW SQL機能を使用して集計クエリを構築
	query := models.DrinkRecords.Query(
		sm.Columns(
			sqlite.Raw("COALESCE(SUM(amount_ml * alcohol_percentage * 0.8), 0) as total_alcohol"),
			sqlite.Raw("COALESCE(SUM(amount_ml), 0) as total_ml"),
		),
		models.SelectWhere.DrinkRecords.UserID.EQ(int32(userID)),
		sm.Where(sqlite.Raw("recorded_at >= ? AND recorded_at < ?", startOfDay.Format("2006-01-02 15:04:05.999999999 -0700 MST"), endOfDay.Format("2006-01-02 15:04:05.999999999 -0700 MST"))),
	)

	sqlQuery, args, err := query.Build(ctx)
	if err != nil {
		return 0, 0, err
	}


	var totalAlcohol sql.NullFloat64
	var totalMl sql.NullInt64

	err = ds.db.QueryRowContext(ctx, sqlQuery, args...).Scan(&totalAlcohol, &totalMl)
	if err != nil {
		return 0, 0, err
	}

	return totalAlcohol.Float64, totalMl.Int64, nil
}

func (ds *DrinkService) DeleteTodayDrinks(ctx context.Context, userID int64) (int, error) {
	// Use datetime range for today
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Use raw SQL for deletion since Bob's delete API is complex
	deleteSQL := "DELETE FROM drink_records WHERE user_id = ? AND recorded_at >= ? AND recorded_at < ?"
	result, err := ds.db.ExecContext(ctx, deleteSQL, 
		int32(userID), 
		startOfDay.Format("2006-01-02 15:04:05.999999999 -0700 MST"), 
		endOfDay.Format("2006-01-02 15:04:05.999999999 -0700 MST"))

	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}
