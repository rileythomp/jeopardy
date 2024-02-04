package jeopardy

import (
	"context"
	"time"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
)

const (
	pickTimeout               = 60 * time.Second
	buzzTimeout               = 60 * time.Second
	defaultAnsTimeout         = 60 * time.Second
	dailyDoubleAnsTimeout     = 60 * time.Second
	finalJeopardyAnsTimeout   = 60 * time.Second
	voteTimeout               = 60 * time.Second
	dailyDoubleWagerTimeout   = 60 * time.Second
	finalJeopardyWagerTimeout = 60 * time.Second

	// pickTimeout               = 2 * time.Second
	// buzzTimeout               = 2 * time.Second
	// defaultAnsTimeout         = 10 * time.Second
	// dailyDoubleAnsTimeout     = 10 * time.Second
	// finalJeopardyAnsTimeout   = 10 * time.Second
	// voteTimeout               = 2 * time.Second
	// dailyDoubleWagerTimeout   = 30 * time.Second
	// finalJeopardyWagerTimeout = 30 * time.Second
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
				log.Errorf("Unexpected error after timeout for player %s: %s\n", player.Name, err)
			}
			return
		}
	}()
}

func (g *Game) startPickTimeout(player *Player) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelPickTimeout = cancel
	g.startTimeout(ctx, pickTimeout, &Player{}, func(_ *Player) error {
		catIdx, valIdx := g.firstAvailableQuestion()
		return g.processPick(player, catIdx, valIdx)
	})
}

func (g *Game) startBuzzTimeout(player *Player) {
	ctx, cancel := context.WithCancel(context.Background())
	g.StartBuzzCountdown = true
	g.cancelBuzzTimeout = cancel
	g.startTimeout(ctx, buzzTimeout, &Player{}, func(_ *Player) error {
		g.skipQuestion()
		return nil
	})
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
	go g.startTimeout(ctx, answerTimeout, player, func(player *Player) error {
		if g.Round == FinalRound {
			return g.processFinalRoundAns(player, false, "answer-timeout")
		}
		g.nextQuestion(player, false)
		return nil
	})
}

func (g *Game) startVoteTimeout(player *Player) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelVoteTimeout = cancel
	g.startTimeout(ctx, voteTimeout, &Player{}, func(_ *Player) error {
		g.nextQuestion(g.LastToAnswer, g.AnsCorrectness)
		return nil
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
