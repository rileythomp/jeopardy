package jeopardy

import (
	"time"

	"github.com/google/uuid"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
)

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

func (g *Game) saveGameAnalytics() {
	fr, frAns, frCorr := getRoundAnalytics(g.FirstRound)
	sr, srAns, srCorr := getRoundAnalytics(g.SecondRound)
	if err := g.questionDB.SaveGameAnalytics(uuid.New(), time.Now().Unix(), fr, frAns, frCorr, sr, srAns, srCorr); err != nil {
		log.Errorf("Error saving game analytics: %s", err.Error())
	}
}

func getRoundAnalytics(round []Category) ([]AnalyticsCategory, int, int) {
	analyticsRound := []AnalyticsCategory{}
	roundAnswers, roundCorrects := 0, 0
	for _, category := range round {
		c := AnalyticsCategory{Title: category.Title}
		for _, question := range category.Questions {
			q := AnalyticsQuestion{}
			if len(question.Answers) > 0 {
				roundAnswers++
			}
			seenCorr := false
			for _, ans := range question.Answers {
				if ans.Correct && !seenCorr {
					seenCorr = true
					roundCorrects++
				}
				answer := AnalyticsAnswer{
					PlayerID:    ans.Player.id(),
					Answer:      ans.Answer,
					Correct:     ans.Correct,
					HasDisputed: ans.HasDisputed,
					Overturned:  ans.Overturned,
				}
				q.Answers = append(q.Answers, answer)
			}
			c.Question = append(c.Question, q)
		}
		analyticsRound = append(analyticsRound, c)
	}
	return analyticsRound, roundAnswers, roundCorrects
}