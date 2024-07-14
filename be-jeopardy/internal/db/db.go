package db

import (
	"context"
	_ "embed"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type (
	Question struct {
		Round        int      `json:"round"`
		Value        int      `json:"value"`
		Category     string   `json:"category"`
		Comments     string   `json:"comments"`
		Clue         string   `json:"question"`
		Answer       string   `json:"-"`
		Alternatives []string `json:"-"`
	}

	Category struct {
		Name    string `json:"category"`
		Round   int    `json:"round"`
		AirDate string `json:"airDate"`
	}

	JeopardyDB struct {
		pool *pgxpool.Pool
	}
)

func NewJeopardyDB(ctx context.Context) (*JeopardyDB, error) {
	poolConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return &JeopardyDB{}, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return &JeopardyDB{}, err
	}
	return &JeopardyDB{pool: pool}, nil
}

func (db *JeopardyDB) Close() {
	db.pool.Close()
}

//go:embed sql/get_questions.sql
var getQuestions string

func (db *JeopardyDB) GetQuestions(ctx context.Context, frCategories, srCategories int) ([]Question, error) {
	rows, err := db.pool.Query(ctx, getQuestions, frCategories, srCategories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := []Question{}
	for rows.Next() {
		var q Question
		err := rows.Scan(&q.Round, &q.Value, &q.Category, &q.Comments, &q.Clue, &q.Answer, &q.Alternatives)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	return questions, nil
}

//go:embed sql/get_category_questions.sql
var getCategoryQuestions string

func (db *JeopardyDB) GetCategoryQuestions(ctx context.Context, category Category) ([]Question, error) {
	rows, err := db.pool.Query(ctx, getCategoryQuestions, category.Name, category.AirDate, category.Round)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := []Question{}
	for rows.Next() {
		var q Question
		err := rows.Scan(&q.Round, &q.Value, &q.Category, &q.Comments, &q.Clue, &q.Answer, &q.Alternatives)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	return questions, nil
}

//go:embed sql/add_alternatives.sql
var addAlternative string

func (db *JeopardyDB) AddAlternative(ctx context.Context, alternative, answer string) error {
	_, err := db.pool.Exec(ctx, addAlternative, alternative, answer)
	return err
}

//go:embed sql/add_incorrect.sql
var addIncorrect string

func (db *JeopardyDB) AddIncorrect(ctx context.Context, incorrect, clue string) error {
	_, err := db.pool.Exec(ctx, addIncorrect, incorrect, clue)
	return err
}

//go:embed sql/search_categories.sql
var searchCategories string

func (db *JeopardyDB) SearchCategories(ctx context.Context, query, start string, secondRound int) ([]Category, error) {
	rows, err := db.pool.Query(ctx, searchCategories, query, secondRound, start)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.Name, &category.Round, &category.AirDate); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

//go:embed sql/increment_player_games.sql
var incrementPlayerGames string

func (db *JeopardyDB) IncrementPlayerGames(ctx context.Context, email string, win, points, answered, correct int) error {
	_, err := db.pool.Exec(ctx, incrementPlayerGames, email, win, points, answered, correct)
	return err
}
