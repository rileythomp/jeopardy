package db

import (
	"context"
	_ "embed"
	"os"

	"github.com/jackc/pgx/v5"
)

type (
	Question struct {
		Round    int    `json:"round"`
		Value    int    `json:"value"`
		Category string `json:"category"`
		Comments string `json:"comments"`
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}

	QuestionDB struct {
		Conn *pgx.Conn
	}
)

func NewQuestionDB() (*QuestionDB, error) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return &QuestionDB{}, err
	}
	return &QuestionDB{Conn: conn}, nil
}

//go:embed sql/get_questions.sql
var getQuestions string

func (db *QuestionDB) GetQuestions() ([]Question, error) {
	rows, err := db.Conn.Query(context.Background(), getQuestions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	questions := []Question{}
	for rows.Next() {
		var q Question
		err := rows.Scan(&q.Round, &q.Value, &q.Category, &q.Comments, &q.Question, &q.Answer)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	return questions, nil
}

func (db *QuestionDB) Close() error {
	return db.Conn.Close(context.Background())
}
