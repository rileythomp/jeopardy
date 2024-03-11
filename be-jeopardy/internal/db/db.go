package db

import (
	"context"
	_ "embed"
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

func (db *JeopardyDB) GetQuestions() ([]Question, error) {
	rows, err := db.pool.Query(context.Background(), getQuestions)
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

type AnalyticsAnswer struct {
	PlayerID    string `json:"playerId"`
	Answer      string `json:"answer"`
	Correct     bool   `json:"correct"`
	HasDisputed bool   `json:"hasDisputed"`
	Overturned  bool   `json:"overturned"`
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
	var (
		gamesPlayed         int
		firstRoundAnsRate   int
		firstRoundCorrRate  int
		firstRoundScore     int
		secondRoundAnsRate  int
		secondRoundCorrRate int
		secondRoundScore    int
	)
	err := db.pool.QueryRow(context.Background(), getAnalytics).Scan(
		&gamesPlayed,
		&firstRoundAnsRate,
		&firstRoundCorrRate,
		&firstRoundScore,
		&secondRoundAnsRate,
		&secondRoundCorrRate,
		&secondRoundScore,
	)
	if err != nil {
		return nil, err
	}
	return struct {
		GamesPlayed         int `json:"gamesPlayed"`
		FirstRoundAnsRate   int `json:"firstRoundAnsRate"`
		FirstRoundCorrRate  int `json:"firstRoundCorrRate"`
		FirstRoundScore     int `json:"firstRoundScore"`
		SecondRoundAnsRate  int `json:"secondRoundAnsRate"`
		SecondRoundCorrRate int `json:"secondRoundCorrRate"`
		SecondRoundScore    int `json:"secondRoundScore"`
	}{
		GamesPlayed:         gamesPlayed,
		FirstRoundAnsRate:   firstRoundAnsRate,
		FirstRoundCorrRate:  firstRoundCorrRate,
		FirstRoundScore:     firstRoundScore,
		SecondRoundAnsRate:  secondRoundAnsRate,
		SecondRoundCorrRate: secondRoundCorrRate,
		SecondRoundScore:    secondRoundScore,
	}, nil
}
