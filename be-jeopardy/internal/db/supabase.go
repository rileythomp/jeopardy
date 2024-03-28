package db

import (
	"context"
	_ "embed"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	User struct {
		ID          string    `json:"id"`
		Email       string    `json:"email"`
		DisplayName string    `json:"displayName"`
		ImgUrl      string    `json:"imgUrl"`
		Public      bool      `json:"public"`
		CreatedAt   time.Time `json:"createdAt"`
	}

	SupabaseDB struct {
		pool *pgxpool.Pool
	}
)

func NewSupabaseDB() (*SupabaseDB, error) {
	poolConfig, err := pgxpool.ParseConfig(os.Getenv("SUPABASE_URL"))
	if err != nil {
		return &SupabaseDB{}, err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return &SupabaseDB{}, err
	}
	return &SupabaseDB{pool: pool}, nil
}

func (db *SupabaseDB) Close() {
	db.pool.Close()
}

//go:embed sql/get_user_by_name.sql
var getUserByName string

func (db *SupabaseDB) GetUserByName(ctx context.Context, name string) (User, error) {
	row := db.pool.QueryRow(ctx, getUserByName, name)
	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.ImgUrl,
		&user.Public,
		&user.CreatedAt,
	)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

//go:embed sql/get_user_by_email.sql
var getUserByEmail string

func (db *SupabaseDB) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := db.pool.QueryRow(ctx, getUserByEmail, email)
	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.ImgUrl,
		&user.Public,
		&user.CreatedAt,
	)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
