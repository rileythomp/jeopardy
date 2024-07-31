package jeopardy

import (
	"context"
	"fmt"
	"strconv"
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

func (g *Game) saveGameAnalytics(ctx context.Context) {
	if !g.Penalty {
		return
	}
	fr, sr := getRoundAnalytics(g.FirstRound), getRoundAnalytics(g.SecondRound)
	if *fr.Answers == 0 || *sr.Answers == 0 {
		// players likely left the game and let it play out so just ignore it
		return
	}
	fr.Score, sr.Score = &g.FirstRoundScore, &g.SecondRoundScore
	if !g.FullGame {
		sr = db.AnalyticsRound{}
	}
	if err := g.jeopardyDB.SaveGameAnalytics(ctx, uuid.New(), time.Now().Unix(), fr, sr); err != nil {
		log.Errorf("Error saving game analytics: %s", err.Error())
	}
	for _, player := range g.Players {
		if !player.isBot() && player.email() != "" {
			wins := 0
			if g.isWinner(player.score()) {
				wins = 1
			}
			answers, correct := g.answersFor(player)
			if err := g.jeopardyDB.IncrementPlayerGames(ctx, player.email(), wins, player.score(), answers, correct); err != nil {
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

func GetAnalytics(ctx context.Context) (any, error) {
	analytics, err := analyticsDB.GetAnalytics(ctx)
	if err != nil {
		log.Errorf("Error getting game analytics: %s", err.Error())
		return nil, err
	}
	return analytics, nil
}

func GetPlayerAnalytics(ctx context.Context, email string) (db.PlayerAnalytics, error) {
	analytics, err := analyticsDB.GetPlayerAnalytics(ctx, email)
	if err != nil {
		log.Errorf("Error getting player analytics: %s", err.Error())
		return db.PlayerAnalytics{}, err
	}
	return analytics, nil
}

var userCache = map[string]db.User{}

func GetLeaderboard(ctx context.Context, leaderboardType string) ([]*db.LeaderboardUser, error) {
	leaderboard, err := analyticsDB.GetLeaderboard(ctx, leaderboardType)
	if err != nil {
		log.Errorf("Error getting leaderboard: %s", err.Error())
		return nil, err
	}
	for _, user := range leaderboard {
		user.WinRate, _ = strconv.ParseFloat(fmt.Sprintf("%.1f", 100*user.WinRate), 64)
		user.CorrectRate, _ = strconv.ParseFloat(fmt.Sprintf("%.1f", 100*user.CorrectRate), 64)
		if u, ok := userCache[user.Email]; ok {
			user.User = u
		} else {
			user.User, err = supabase.GetUserByEmail(ctx, user.Email)
			if err != nil {
				log.Errorf("Error getting user: %s", err.Error())
				return nil, err
			}
		}
		userCache[user.Email] = user.User
	}
	return leaderboard, nil
}
