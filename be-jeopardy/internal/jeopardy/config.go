package jeopardy

import "time"

type GameConfig struct {
	FullGame bool `json:"fullGame"`
	Penalty  bool `json:"penalty"`

	pickTimeout        time.Duration
	buzzTimeout        time.Duration
	answerTimeout      time.Duration
	finalAnswerTimeout time.Duration
	voteTimeout        time.Duration
	disputeTimeout     time.Duration
	wagerTimeout       time.Duration
	finalWagerTimeout  time.Duration
}
