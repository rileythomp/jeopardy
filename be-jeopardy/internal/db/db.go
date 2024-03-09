package db

import (
	"context"
	_ "embed"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
		Conn *pgx.Conn
	}
)

func NewJeopardyDB() (*JeopardyDB, error) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return &JeopardyDB{}, err
	}
	return &JeopardyDB{Conn: conn}, nil
}

func (db *JeopardyDB) Close() error {
	return db.Conn.Close(context.Background())
}

//go:embed sql/get_questions.sql
var getQuestions string

func (db *JeopardyDB) GetQuestions() ([]Question, error) {
	rows, err := db.Conn.Query(context.Background(), getQuestions)
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
	_, err := db.Conn.Exec(context.Background(), addAlternative, alternative, answer)
	return err
}

//go:embed sql/save_game_analytics.sql
var saveGameAnalytics string

func (db *JeopardyDB) SaveGameAnalytics(gameID uuid.UUID, createdAt int64, firstRound any, frAns, frCorr int, secondRound any, srAns, srCorr int) error {
	_, err := db.Conn.Exec(context.Background(), saveGameAnalytics, gameID, createdAt, firstRound, frAns, frCorr, secondRound, srAns, srCorr)
	return err
}
