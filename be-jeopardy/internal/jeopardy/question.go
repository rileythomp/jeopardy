package jeopardy

import (
	"context"
	"math/rand/v2"
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
		Title     string      `json:"title"`
		Questions []*Question `json:"questions"`
	}

	Answer struct {
		Player      GamePlayer `json:"player"`
		Answer      string     `json:"answer"`
		Correct     bool       `json:"correct"`
		HasDisputed bool       `json:"hasDisputed"`
		Overturned  bool       `json:"overturned"`
		Bot         bool       `json:"bot"`
	}

	Question struct {
		db.Question
		CanChoose   bool `json:"canChoose"`
		DailyDouble bool `json:"-"`

		Answers     []*Answer `json:"answers"`
		CurAns      *Answer   `json:"curAns"`
		CurDisputed *Answer   `json:"curDisputed"`
	}
)

func (q *Question) checkAnswer(ans string) bool {
	for _, corr := range q.Alternatives {
		ans, corr = strings.ToLower(ans), strings.ToLower(corr)
		dist := levenshtein.ComputeDistance(ans, corr)
		if len(corr) <= 5 && dist <= 0 {
			return true
		} else if 5 < len(corr) && len(corr) <= 7 && dist <= 1 {
			return true
		} else if 7 < len(corr) && len(corr) <= 9 && dist <= 2 {
			return true
		} else if 9 < len(corr) && len(corr) <= 12 && dist <= 3 {
			return true
		} else if 12 < len(corr) && len(corr) <= 15 && dist <= 4 {
			return true
		} else if 15 < len(corr) && dist <= 5 {
			return true
		}
	}
	return false
}

func (q *Question) equal(q0 *Question) bool {
	return q.Clue == q0.Clue && q.Answer == q0.Answer
}

func (g *Game) setQuestions(ctx context.Context) error {
	g.FirstRound = []Category{}
	g.SecondRound = []Category{}
	g.FinalQuestion = &Question{}
	g.CurQuestion = &Question{}
	g.OfficialAnswer = ""

	questions := []db.Question{}

	categories := append(g.FirstRoundCategories, g.SecondRoundCategories...)
	for _, category := range categories {
		categoryQuestions, err := g.jeopardyDB.GetCategoryQuestions(ctx, category)
		if err != nil {
			return err
		}
		questions = append(questions, categoryQuestions...)
	}

	randomQuestions, err := g.jeopardyDB.GetQuestions(ctx, 6-len(g.FirstRoundCategories), 6-len(g.SecondRoundCategories))
	if err != nil {
		return err
	}
	questions = append(questions, randomQuestions...)

	category := Category{}
	for i, q := range questions {
		question := &Question{Question: q}
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
	tIdx := rand.IntN(numCategories)
	qIdx := 0
	num := rand.IntN(10000)
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
	tIdx := rand.IntN(numCategories)
	qIdx := 0
	num := rand.IntN(10000)
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

func (g *Game) nextQuestionInCategory() (int, int) {
	curRound := g.FirstRound
	if g.Round == SecondRound {
		curRound = g.SecondRound
	}
	for catIdx, category := range curRound {
		if category.Title == g.CurQuestion.Category {
			for valIdx, question := range category.Questions {
				if question.CanChoose {
					return catIdx, valIdx
				}
			}
		}
	}
	return g.firstAvailableQuestion()
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
