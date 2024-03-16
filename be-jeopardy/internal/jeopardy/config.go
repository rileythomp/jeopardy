package jeopardy

import (
	"fmt"
)

type GameConfig struct {
	FullGame bool `json:"fullGame"`
	Penalty  bool `json:"penalty"`
	Bots     int  `json:"bots"`

	PickTimeout        int `json:"pickTimeout"`
	BuzzTimeout        int `json:"buzzTimeout"`
	AnswerTimeout      int `json:"answerTimeout"`
	FinalAnswerTimeout int `json:"finalAnswerTimeout"`
	VoteTimeout        int `json:"voteTimeout"`
	DisputeTimeout     int `json:"disputeTimeout"`
	WagerTimeout       int `json:"wagerTimeout"`
	FinalWagerTimeout  int `json:"finalWagerTimeout"`
}

func NewConfig(
	fullGame, penalty bool, bots int,
	pickTimeout, buzzTimeout, answerTimeout, voteTimeout, wagerTimeout int,
) (GameConfig, error) {
	if bots < 0 || bots > 2 {
		return GameConfig{}, fmt.Errorf("Bots must be between 0 and 2, got: %d", bots)
	}
	if pickTimeout < 3 || pickTimeout > 60 {
		return GameConfig{}, fmt.Errorf("Pick timeout must be between 3 and 60 seconds, got: %d", pickTimeout)
	}
	if buzzTimeout < 3 || buzzTimeout > 60 {
		return GameConfig{}, fmt.Errorf("Buzz timeout must be between 3 and 60 seconds, got: %d", buzzTimeout)
	}
	if answerTimeout < 3 || answerTimeout > 60 {
		return GameConfig{}, fmt.Errorf("Answer timeout must be between 3 and 60 seconds, got: %d", answerTimeout)
	}
	if voteTimeout < 3 || voteTimeout > 60 {
		return GameConfig{}, fmt.Errorf("Vote timeout must be between 3 and 60 seconds, got: %d", voteTimeout)
	}
	if wagerTimeout < 3 || wagerTimeout > 60 {
		return GameConfig{}, fmt.Errorf("Wager timeout must be between 3 and 60 seconds, got: %d", wagerTimeout)
	}
	return GameConfig{
		FullGame:           fullGame,
		Penalty:            penalty,
		Bots:               bots,
		PickTimeout:        pickTimeout,
		BuzzTimeout:        buzzTimeout,
		AnswerTimeout:      answerTimeout,
		FinalAnswerTimeout: 30,
		VoteTimeout:        voteTimeout,
		DisputeTimeout:     60,
		WagerTimeout:       wagerTimeout,
		FinalWagerTimeout:  30,
	}, nil
}

// pickTimeout:        30 * time.Second,
// buzzTimeout:        30 * time.Second,
// answerTimeout:      15 * time.Second,
// finalAnswerTimeout: 30 * time.Second,
// voteTimeout:        10 * time.Second,
// disputeTimeout:     60 * time.Second,
// wagerTimeout:       30 * time.Second,
// finalWagerTimeout:  30 * time.Second,
