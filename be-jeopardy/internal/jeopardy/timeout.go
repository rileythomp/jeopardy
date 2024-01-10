package jeopardy

import (
	"context"
	"log"
	"time"
)

const (
	// pickQuestionTimeout       = 10 * time.Second
	// buzzInTimeout             = 10 * time.Second
	// defaultAnsTimeout         = 10 * time.Second
	// dailyDoubleAnsTimeout     = 10 * time.Second
	// finalJeopardyAnsTimeout   = 10 * time.Second
	// confirmAnsTimeout         = 10 * time.Second
	// dailyDoubleWagerTimeout   = 10 * time.Second
	// finalJeopardyWagerTimeout = 10 * time.Second

	pickQuestionTimeout       = 2 * time.Second
	buzzInTimeout             = 2 * time.Second
	defaultAnsTimeout         = 10 * time.Second
	dailyDoubleAnsTimeout     = 10 * time.Second
	finalJeopardyAnsTimeout   = 10 * time.Second
	confirmAnsTimeout         = 2 * time.Second
	dailyDoubleWagerTimeout   = 10 * time.Second
	finalJeopardyWagerTimeout = 10 * time.Second
)

func (g *Game) startTimeout(ctx context.Context, timeout time.Duration, player *Player, processTimeout func(player *Player) error) {
	go func() {
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
		defer timeoutCancel()
		select {
		case <-ctx.Done():
			return
		case <-timeoutCtx.Done():
			if err := processTimeout(player); err != nil {
				log.Printf("Unexpected error after timeout for player %s: %s\n", player.Name, err)
				panic("error processing a timeout")
			}
			return
		}
	}()
}

func (g *Game) startPickTimeout(player *Player) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelPickTimeout = cancel
	g.startTimeout(ctx, pickQuestionTimeout, &Player{}, func(_ *Player) error {
		topicIdx, valIdx := g.firstAvailableQuestion()
		return g.processPick(player, topicIdx, valIdx)
	})
}

func (g *Game) startBuzzTimeout(player *Player) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelBuzzTimeout = cancel
	g.startTimeout(ctx, buzzInTimeout, &Player{}, func(_ *Player) error { return g.skipQuestion() })
}

func (g *Game) startAnswerTimeout(player *Player) {
	ctx, cancel := context.WithCancel(context.Background())
	player.cancelAnswerTimeout = cancel
	answerTimeout := defaultAnsTimeout
	if g.CurQuestion.DailyDouble {
		answerTimeout = dailyDoubleAnsTimeout
	} else if g.Round == FinalRound {
		answerTimeout = finalJeopardyAnsTimeout
	}
	go g.startTimeout(ctx, answerTimeout, player, g.processAnswerTimeout)
}

func (g *Game) startConfirmationTimeout(player *Player) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelConfirmationTimeout = cancel
	g.startTimeout(ctx, confirmAnsTimeout, &Player{}, func(_ *Player) error {
		return g.nextQuestion(g.LastToAnswer, g.AnsCorrectness)
	})
}

func (g *Game) startWagerTimeout(player *Player) {
	ctx, cancel := context.WithCancel(context.Background())
	player.cancelWagerTimeout = cancel
	wagerTimeout := dailyDoubleWagerTimeout
	if g.Round == FinalRound {
		wagerTimeout = finalJeopardyWagerTimeout
	}
	g.startTimeout(ctx, wagerTimeout, player, func(player *Player) error {
		wager := 5
		if g.Round == FinalRound {
			wager = 0
		}
		return g.processWager(player, wager)
	})
}
