package jeopardy

import (
	"math/rand"
	"strings"

	"github.com/agnivade/levenshtein"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
)

const (
	numCategories = 6
	numQuestions  = 5
)

type (
	Category struct {
		Title     string     `json:"title"`
		Questions []Question `json:"questions"`
	}

	Question struct {
		db.Question
		CanChoose   bool `json:"canChoose"`
		DailyDouble bool `json:"dailyDouble"`
	}
)

func (q *Question) checkAnswer(ans string) bool {
	for _, corr := range q.Alternatives {
		ans, corr = strings.ToLower(ans), strings.ToLower(corr)
		dist := levenshtein.ComputeDistance(ans, corr)
		if len(corr) <= 5 && dist <= 0 {
			return true
		} else if len(corr) <= 7 && dist <= 1 {
			return true
		} else if len(corr) <= 9 && dist <= 2 {
			return true
		} else if len(corr) <= 12 && dist <= 3 {
			return true
		} else if len(corr) <= 15 && dist <= 4 {
			return true
		} else if len(corr) > 15 && dist <= 5 {
			return true
		}
	}
	return false
}

func (q *Question) equal(q0 Question) bool {
	return q.Clue == q0.Clue && q.Answer == q0.Answer
}

func (g *Game) setQuestions() error {
	g.FirstRound = []Category{}
	g.SecondRound = []Category{}
	g.FinalQuestion = Question{}
	g.CurQuestion = Question{}
	questions, err := g.questionDB.GetQuestions()
	if err != nil {
		return err
	}
	category := Category{}
	for i, q := range questions {
		question := Question{Question: q}
		if question.Round == 3 {
			g.FinalQuestion = question
			continue
		}
		question.CanChoose = true
		category.Questions = append(category.Questions, question)
		if i%numQuestions == (numQuestions - 1) {
			category.Title = question.Category
			if question.Round == 1 {
				g.FirstRound = append(g.FirstRound, category)
			} else {
				g.SecondRound = append(g.SecondRound, category)
			}
			category = Category{}
		}
	}
	g.setDailyDoubles()
	return nil
}

func (g *Game) setDailyDoubles() {
	// based on daily_double_occurrence_bounds.sql
	g.setFirstRoundDailyDouble()
	g.setSecondRoundDailyDouble()
	g.setSecondRoundDailyDouble()
}

func (g *Game) setFirstRoundDailyDouble() {
	tIdx := rand.Intn(numCategories)
	qIdx := 0
	num := rand.Intn(10000)
	if num < 15 {
		qIdx = 0
	} else if num < 1150 {
		qIdx = 1
	} else if num < 3916 {
		qIdx = 2
	} else if num < 7409 {
		qIdx = 3
	} else {
		qIdx = 4
	}
	g.FirstRound[tIdx].Questions[qIdx].DailyDouble = true
}

func (g *Game) setSecondRoundDailyDouble() {
	tIdx := rand.Intn(numCategories)
	qIdx := 0
	num := rand.Intn(10000)
	if num < 15 {
		qIdx = 0
	} else if num < 1524 {
		qIdx = 1
	} else if num < 4682 {
		qIdx = 2
	} else if num < 8220 {
		qIdx = 3
	} else {
		qIdx = 4
	}
	g.SecondRound[tIdx].Questions[qIdx].DailyDouble = true
}

func (g *Game) firstAvailableQuestion() (int, int) {
	curRound := g.FirstRound
	if g.Round == SecondRound {
		curRound = g.SecondRound
	}
	for valIdx := 0; valIdx < numQuestions; valIdx++ {
		for catIdx := 0; catIdx < numCategories; catIdx++ {
			if curRound[catIdx].Questions[valIdx].CanChoose {
				return catIdx, valIdx
			}
		}

	}
	return -1, -1
}

func (g *Game) disableQuestion() {
	for i, category := range g.FirstRound {
		for j, q := range category.Questions {
			if q.equal(g.CurQuestion) {
				g.FirstRound[i].Questions[j].CanChoose = false
			}
		}
	}
	for i, category := range g.SecondRound {
		for j, q := range category.Questions {
			if q.equal(g.CurQuestion) {
				g.SecondRound[i].Questions[j].CanChoose = false
			}
		}
	}
}
