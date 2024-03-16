package jeopardy

import (
	"fmt"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/db"
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

	FirstRoundCategories  []db.Category `json:"firstRoundCategories"`
	SecondRoundCategories []db.Category `json:"secondRoundCategories"`
}

func NewConfig(
	fullGame, penalty bool, bots int,
	pickTimeout, buzzTimeout, answerTimeout, voteTimeout, wagerTimeout int,
	firstRoundCategories, secondRoundCategories []db.Category,
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
	if len(firstRoundCategories) > 6 {
		return GameConfig{}, fmt.Errorf("First round cannot have more than 6 categories, got: %d", len(firstRoundCategories))
	}
	if len(secondRoundCategories) > 6 {
		return GameConfig{}, fmt.Errorf("Second round cannot have more than 6 categories, got: %d", len(secondRoundCategories))
	}
	return GameConfig{
		FullGame:              fullGame,
		Penalty:               penalty,
		Bots:                  bots,
		PickTimeout:           pickTimeout,
		BuzzTimeout:           buzzTimeout,
		AnswerTimeout:         answerTimeout,
		FinalAnswerTimeout:    30,
		VoteTimeout:           voteTimeout,
		DisputeTimeout:        60,
		WagerTimeout:          wagerTimeout,
		FinalWagerTimeout:     30,
		FirstRoundCategories:  firstRoundCategories,
		SecondRoundCategories: secondRoundCategories,
	}, nil
}
