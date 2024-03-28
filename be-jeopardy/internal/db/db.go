package db

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/google/uuid"
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

func NewJeopardyDB() (*JeopardyDB, error) {
	poolConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return &JeopardyDB{}, err
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
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

func (db *JeopardyDB) GetQuestions(firstRoundCategories, secondRoundCategories int) ([]Question, error) {
	rows, err := db.pool.Query(context.Background(), getQuestions, firstRoundCategories, secondRoundCategories)
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

func (db *JeopardyDB) GetCategoryQuestions(category Category) ([]Question, error) {
	rows, err := db.pool.Query(context.Background(), getCategoryQuestions, category.Name, category.AirDate, category.Round)
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

func (db *JeopardyDB) AddAlternative(alternative, answer string) error {
	_, err := db.pool.Exec(context.Background(), addAlternative, alternative, answer)
	return err
}

//go:embed sql/add_incorrect.sql
var addIncorrect string

func (db *JeopardyDB) AddIncorrect(incorrect, clue string) error {
	_, err := db.pool.Exec(context.Background(), addIncorrect, incorrect, clue)
	return err
}

type AnalyticsRound struct {
	Categories []AnalyticsCategory
	Answers    *int
	Correct    *int
	Score      *float64
}

type AnalyticsCategory struct {
	Title    string              `json:"title"`
	Question []AnalyticsQuestion `json:"question"`
}

type AnalyticsQuestion struct {
	Answers []AnalyticsAnswer `json:"answers"`
}

type AnalyticsUser struct {
	User
	Games       int     `json:"games"`
	Wins        int     `json:"wins"`
	Points      int     `json:"points"`
	Answers     int     `json:"answers"`
	Correct     int     `json:"correct"`
	MaxPoints   int     `json:"maxPoints"`
	MaxCorrect  int     `json:"maxCorrect"`
	WinRate     float64 `json:"winRate"`
	CorrectRate float64 `json:"correctRate"`
}

type AnalyticsAnswer struct {
	PlayerID    string `json:"playerId"`
	Answer      string `json:"answer"`
	Correct     bool   `json:"correct"`
	HasDisputed bool   `json:"hasDisputed"`
	Overturned  bool   `json:"overturned"`
	Bot         bool   `json:"bot"`
}

//go:embed sql/save_game_analytics.sql
var saveGameAnalytics string

func (db *JeopardyDB) SaveGameAnalytics(gameID uuid.UUID, createdAt int64, fr AnalyticsRound, sr AnalyticsRound) error {
	_, err := db.pool.Exec(
		context.Background(), saveGameAnalytics, gameID, createdAt,
		fr.Categories, fr.Answers, fr.Correct, fr.Score,
		sr.Categories, sr.Answers, sr.Correct, sr.Score,
	)
	return err
}

//go:embed sql/get_analytics.sql
var getAnalytics string

func (db *JeopardyDB) GetAnalytics() (any, error) {
	a := struct {
		GamesPlayed         int `json:"gamesPlayed"`
		FirstRoundAnsRate   int `json:"firstRoundAnsRate"`
		FirstRoundCorrRate  int `json:"firstRoundCorrRate"`
		FirstRoundScore     int `json:"firstRoundScore"`
		SecondRoundAnsRate  int `json:"secondRoundAnsRate"`
		SecondRoundCorrRate int `json:"secondRoundCorrRate"`
		SecondRoundScore    int `json:"secondRoundScore"`
	}{}
	err := db.pool.QueryRow(context.Background(), getAnalytics).Scan(
		&a.GamesPlayed,
		&a.FirstRoundAnsRate,
		&a.FirstRoundCorrRate,
		&a.FirstRoundScore,
		&a.SecondRoundAnsRate,
		&a.SecondRoundCorrRate,
		&a.SecondRoundScore,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

//go:embed sql/get_leaderboards.sql
var getLeaderboards string

func (db *JeopardyDB) GetLeaderboard(ctx context.Context, leaderboardType string) ([]*AnalyticsUser, error) {
	switch leaderboardType {
	case "win_rate", "wins", "games", "correct_rate", "correct", "answers", "points", "max_points", "max_correct":
	default:
		return nil, fmt.Errorf("invalid leaderboard type: %s", leaderboardType)
	}
	rows, err := db.pool.Query(ctx, fmt.Sprintf(getLeaderboards, leaderboardType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	leaderboard := []*AnalyticsUser{}
	for rows.Next() {
		user := &AnalyticsUser{}
		if err := rows.Scan(
			&user.Email,
			&user.Wins,
			&user.Games,
			&user.WinRate,
			&user.Correct,
			&user.Answers,
			&user.CorrectRate,
			&user.Points,
			&user.MaxPoints,
			&user.MaxCorrect,
		); err != nil {
			return nil, err
		}
		leaderboard = append(leaderboard, user)
	}
	return leaderboard, nil
}

//go:embed sql/get_player_analytics.sql
var getPlayerAnalytics string

func (db *JeopardyDB) GetPlayerAnalytics(email string) (any, error) {
	a := struct {
		Games      int `json:"games"`
		Wins       int `json:"wins"`
		Points     int `json:"points"`
		Answers    int `json:"answers"`
		Correct    int `json:"correct"`
		MaxPoints  int `json:"maxPoints"`
		MaxCorrect int `json:"maxCorrect"`
	}{}
	err := db.pool.QueryRow(context.Background(), getPlayerAnalytics, email).Scan(
		&a.Games,
		&a.Wins,
		&a.Points,
		&a.Answers,
		&a.Correct,
		&a.MaxPoints,
		&a.MaxCorrect,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

//go:embed sql/search_categories.sql
var searchCategories string

func (db *JeopardyDB) SearchCategories(query, start string, secondRound int) ([]Category, error) {
	rows, err := db.pool.Query(context.Background(), searchCategories, query, secondRound, start)
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

func (db *JeopardyDB) IncrementPlayerGames(email string, win, points, answered, correct int) error {
	_, err := db.pool.Exec(context.Background(), incrementPlayerGames, email, win, points, answered, correct)
	return err
}
