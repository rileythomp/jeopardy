package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetQuestions(t *testing.T) {
	t.Run("test getting questions", func(t *testing.T) {
		conn, err := pgx.Connect(context.Background(), "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
		if err != nil {
			t.Fatalf("Error connecting to database: %s", err.Error())
		}
		questionDB := JeopardyDB{Conn: conn}
		questions, err := questionDB.GetQuestions()
		if err != nil {
			t.Fatalf("Error getting questions: %s", err.Error())
		}
		assert.Len(t, questions, 61)
		for i, question := range questions {
			if i < 30 {
				assert.Equal(t, 1, question.Round)
				assert.Equal(t, 200*(i%5+1), question.Value)
			} else if i < 60 {
				assert.Equal(t, 2, question.Round)
				assert.Equal(t, 400*(i%5+1), question.Value)
			} else {
				assert.Equal(t, 3, question.Round)
				assert.Equal(t, 0, question.Value)
			}
		}
	})
}
