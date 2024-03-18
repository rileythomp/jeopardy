package jeopardy

import (
	"fmt"
	"testing"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestPickWager(t *testing.T) {

	tests := []struct {
		score1 int
		score2 int
		score3 int
		want   int
	}{
		{1000, 2000, 5000, 3000},
		{1000, 4000, 5000, 1000},
		{0, 100, 400, 1000},
		{200, 400, 500, 1000},
		{-1000, -2000, 5000, 5000},
		{1000, 5000, 5000, 1000},
		{1000, 5000, 2000, 2000},
		{1000, 5000, 4000, 1000},
		{100, 500, 400, 1000},
		{100, 500, 200, 1000},
		{5000, 5000, 5000, 1000},
		{3000, 5000, 2000, 2000},
		{4000, 5000, 3000, 2000},
		{200, 500, 100, 1000},
		{400, 500, 300, 1000},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			p1, p2, bot := NewPlayer("a"), NewPlayer("b"), NewBot("c")
			p1.addToScore(tc.score1)
			p2.addToScore(tc.score2)
			bot.addToScore(tc.score3)
			players := []GamePlayer{p1, p2, bot}
			wager := bot.pickWager(players, 1000)
			assert.Equal(t, tc.want, wager)
		})
	}
}

func pickQuestion(g *Game, catIdx, valIdx int) {
	g.CurQuestion = g.FirstRound[catIdx].Questions[valIdx]
	g.disableQuestion()
}

func TestPickQuestion(t *testing.T) {
	t.Run("test pick question", func(t *testing.T) {
		questionDB, err := db.NewJeopardyDB()
		if err != nil {
			t.Fatalf("Failed to create questionDB: %s", err)
		}
		config, err := NewConfig(true, true, 0, 30, 30, 30, 30, nil, nil)
		if err != nil {
			t.Fatalf("Failed to create config: %s", err)
		}
		game, err := NewGame(questionDB, config)
		if err != nil {
			t.Fatalf("Failed to create game: %s", err)
		}

		pickQuestion(game, 2, 2)

		for i := 0; i < 5; i++ {
			if i == 2 {
				continue
			}
			catIdx, valIdx := game.nextQuestionInCategory()
			assert.Equal(t, 2, catIdx)
			assert.Equal(t, i, valIdx)
			pickQuestion(game, catIdx, valIdx)
		}

		for i := 0; i < 5; i++ {
			catIdx, valIdx := game.nextQuestionInCategory()
			assert.Equal(t, 0, catIdx)
			assert.Equal(t, i, valIdx)
			pickQuestion(game, catIdx, valIdx)
		}

		for i := 0; i < 5; i++ {
			catIdx, valIdx := game.nextQuestionInCategory()
			assert.Equal(t, 1, catIdx)
			assert.Equal(t, i, valIdx)
			pickQuestion(game, catIdx, valIdx)
		}
	})
}
