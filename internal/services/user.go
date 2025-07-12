package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/aberyotaro/drink-tracker/models"
)

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

func (us *UserService) GetOrCreateUser(ctx context.Context, slackUserID, slackTeamID string) (*models.User, error) {
	user, err := us.GetUserBySlackID(ctx, slackUserID)
	if err == nil {
		return user, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	return us.CreateUser(ctx, slackUserID, slackTeamID)
}

func (us *UserService) GetUserBySlackID(ctx context.Context, slackUserID string) (*models.User, error) {
	// Bobのクエリビルダーを使用してSQLとArgsを生成
	query := models.Users.Query(
		models.SelectWhere.Users.SlackUserID.EQ(slackUserID),
	)

	sqlQuery, args, err := query.Build(ctx)
	if err != nil {
		return nil, err
	}

	// 生成されたSQLを標準のsql.DBで実行
	var user models.User
	err = us.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&user.ID, &user.SlackUserID, &user.SlackTeamID,
		&user.DailyLimitML, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (us *UserService) CreateUser(ctx context.Context, slackUserID, slackTeamID string) (*models.User, error) {
	now := time.Now()

	// BobのInsertクエリビルダーを使用
	setter := &models.UserSetter{
		SlackUserID:  &slackUserID,
		SlackTeamID:  &slackTeamID,
		DailyLimitML: &sql.Null[int32]{V: 40000, Valid: true},
		CreatedAt:    &sql.Null[time.Time]{V: now, Valid: true},
		UpdatedAt:    &sql.Null[time.Time]{V: now, Valid: true},
	}

	insertQuery := models.Users.Insert(setter)
	sqlQuery, args, err := insertQuery.Build(ctx)
	if err != nil {
		return nil, err
	}

	result, err := us.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           int32(id),
		SlackUserID:  slackUserID,
		SlackTeamID:  slackTeamID,
		DailyLimitML: sql.Null[int32]{V: 40000, Valid: true},
		CreatedAt:    sql.Null[time.Time]{V: now, Valid: true},
		UpdatedAt:    sql.Null[time.Time]{V: now, Valid: true},
	}

	return user, nil
}
