package jeopardy

import "time"

type GameConfig struct {
	FullGame bool `json:"fullGame"`
	Penalty  bool `json:"penalty"`
	Bots     int  `json:"bots"`

	pickTimeout        time.Duration
	buzzTimeout        time.Duration
	answerTimeout      time.Duration
	finalAnswerTimeout time.Duration
	voteTimeout        time.Duration
	disputeTimeout     time.Duration
	wagerTimeout       time.Duration
	finalWagerTimeout  time.Duration
}

func NewConfig(fullGame, penalty bool, bots int) GameConfig {
	return GameConfig{
		FullGame:           fullGame,
		Penalty:            penalty,
		Bots:               bots,
		pickTimeout:        30 * time.Second,
		buzzTimeout:        30 * time.Second,
		answerTimeout:      15 * time.Second,
		finalAnswerTimeout: 30 * time.Second,
		voteTimeout:        10 * time.Second,
		disputeTimeout:     60 * time.Second,
		wagerTimeout:       30 * time.Second,
		finalWagerTimeout:  30 * time.Second,
	}
}
