package db

import (
	"context"
	_ "embed"
	"os"
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/jackc/pgx/v5"
)

type (
	Question struct {
		Round       int    `json:"round"`
		Value       int    `json:"value"`
		Category    string `json:"category"`
		Comments    string `json:"comments"`
		Question    string `json:"question"`
		Answer      string `json:"answer"`
		AirDate     string `json:"airDate"`
		Notes       string `json:"notes"`
		CanChoose   bool   `json:"canChoose"`
		DailyDouble bool   `json:"dailyDouble"`
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
		err := rows.Scan(&q.Round, &q.Value, &q.Category, &q.Comments, &q.Question, &q.Answer, &q.Comments)
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

func (q *Question) CheckAnswer(ans string) bool {
	ans = strings.ToLower(ans)
	corrAns := strings.ToLower(q.Answer)
	if len(ans) < 5 {
		return ans == corrAns
	} else if len(corrAns) < 7 {
		return levenshtein.ComputeDistance(ans, corrAns) < 2
	} else if len(corrAns) < 9 {
		return levenshtein.ComputeDistance(ans, corrAns) < 3
	} else if len(corrAns) < 11 {
		return levenshtein.ComputeDistance(ans, corrAns) < 4
	} else if len(corrAns) < 13 {
		return levenshtein.ComputeDistance(ans, corrAns) < 5
	} else if len(corrAns) < 15 {
		return levenshtein.ComputeDistance(ans, corrAns) < 6
	}
	return levenshtein.ComputeDistance(ans, corrAns) < 7
}

func (q *Question) Equal(q0 Question) bool {
	return q.Question == q0.Question && q.Answer == q0.Answer
}
