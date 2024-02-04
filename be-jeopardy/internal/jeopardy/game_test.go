package jeopardy

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestSetQuestions(t *testing.T) {
	t.Run("test setting questions", func(t *testing.T) {
		conn, err := pgx.Connect(context.Background(), "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
		if err != nil {
			t.Fatalf("Error connecting to database: %s", err.Error())
		}
		g := Game{questionDB: &db.QuestionDB{Conn: conn}}
		err = g.setQuestions()
		assert.NoError(t, err)
		assert.Len(t, g.FirstRound, 6)
		for _, category := range g.FirstRound {
			assert.Len(t, category.Questions, 5)
			for i, question := range category.Questions {
				assert.Equal(t, ((i%5)+1)*200, question.Value)
			}
		}
		assert.Len(t, g.SecondRound, 6)
		for _, category := range g.SecondRound {
			assert.Len(t, category.Questions, 5)
			for i, question := range category.Questions {
				assert.Equal(t, ((i%5)+1)*400, question.Value)
			}
		}
		assert.Zero(t, g.FinalQuestion.Value)
	})
}
