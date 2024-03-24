package jeopardy

import (
	"time"

	"github.com/google/uuid"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
)

type GameAnalytics struct {
	FirstRoundScore  float64 `json:"firstRoundScore"`
	SecondRoundScore float64 `json:"secondRoundScore"`
}

func (g *Game) isWinner(score int) bool {
	for _, player := range g.Players {
		if player.score() > score {
			return false
		}
	}
	return true
}

func (g *Game) answersFor(player GamePlayer) (int, int) {
	answers, correct := 0, 0
	for _, category := range g.FirstRound {
		for _, question := range category.Questions {
			for _, answer := range question.Answers {
				if answer.Player.id() == player.id() {
					answers++
					if answer.Correct {
						correct++
					}
				}
			}
		}
	}
	for _, category := range g.SecondRound {
		for _, question := range category.Questions {
			for _, answer := range question.Answers {
				if answer.Player.id() == player.id() {
					answers++
					if answer.Correct {
						correct++
					}
				}
			}
		}
	}
	return answers, correct
}

func (g *Game) saveGameAnalytics() {
	if !g.Penalty {
		return
	}
	fr, sr := getRoundAnalytics(g.FirstRound), getRoundAnalytics(g.SecondRound)
	fr.Score, sr.Score = &g.FirstRoundScore, &g.SecondRoundScore
	if !g.FullGame {
		sr = db.AnalyticsRound{}
	}
	if err := g.jeopardyDB.SaveGameAnalytics(uuid.New(), time.Now().Unix(), fr, sr); err != nil {
		log.Errorf("Error saving game analytics: %s", err.Error())
	}
	for _, player := range g.Players {
		if !player.isBot() && player.email() != "" {
			wins := 0
			if g.isWinner(player.score()) {
				wins = 1
			}
			answers, correct := g.answersFor(player)
			if err := g.jeopardyDB.IncrementPlayerGames(player.email(), wins, player.score(), answers, correct); err != nil {
				log.Errorf("Error incrementing player game count: %s", err.Error())
			}
		}
	}
}

func getRoundAnalytics(round []Category) db.AnalyticsRound {
	categories := []db.AnalyticsCategory{}
	answers, correct := 0, 0
	for _, category := range round {
		c := db.AnalyticsCategory{Title: category.Title}
		for _, question := range category.Questions {
			q := db.AnalyticsQuestion{}
			seenAns, seenCorr := false, false
			for _, ans := range question.Answers {
				if !seenAns && !ans.Bot {
					seenAns = true
					answers++
				}
				if ans.Correct && !seenCorr && !ans.Bot {
					seenCorr = true
					correct++
				}
				answer := db.AnalyticsAnswer{
					PlayerID:    ans.Player.id(),
					Answer:      ans.Answer,
					Correct:     ans.Correct,
					HasDisputed: ans.HasDisputed,
					Overturned:  ans.Overturned,
					Bot:         ans.Bot,
				}
				q.Answers = append(q.Answers, answer)
			}
			c.Question = append(c.Question, q)
		}
		categories = append(categories, c)
	}
	return db.AnalyticsRound{
		Categories: categories,
		Answers:    &answers,
		Correct:    &correct,
	}
}

func GetAnalytics() (any, error) {
	analytics, err := analyticsDB.GetAnalytics()
	if err != nil {
		log.Errorf("Error getting game analytics: %s", err.Error())
		return nil, err
	}
	return analytics, nil
}

func GetPlayerAnalytics(email string) (any, error) {
	analytics, err := analyticsDB.GetPlayerAnalytics(email)
	if err != nil {
		log.Errorf("Error getting player analytics: %s", err.Error())
		return nil, err
	}
	return analytics, nil
}
