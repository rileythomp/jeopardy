package jeopardy

import (
	"context"
	"time"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
)

const (
	boardIntroTimeout         = 25 * time.Second
	pickTimeout               = 30 * time.Second
	buzzTimeout               = 30 * time.Second
	defaultAnsTimeout         = 30 * time.Second
	dailyDoubleAnsTimeout     = 30 * time.Second
	finalJeopardyAnsTimeout   = 30 * time.Second
	voteTimeout               = 10 * time.Second
	dailyDoubleWagerTimeout   = 30 * time.Second
	finalJeopardyWagerTimeout = 30 * time.Second
)

func (g *Game) startTimeout(ctx context.Context, timeout time.Duration, player GamePlayer, processTimeout func(player GamePlayer) error) {
	go func() {
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), timeout)
		defer timeoutCancel()
		select {
		case <-ctx.Done():
			return
		case <-timeoutCtx.Done():
			if err := processTimeout(player); err != nil {
				log.Errorf("Unexpected error after timeout for player %s: %s\n", player.name(), err)
			}
			return
		}
	}()
}

func (g *Game) startBoardIntroTimeout(player GamePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelBoardIntroTimeout = cancel
	g.startTimeout(ctx, boardIntroTimeout, &Player{}, func(_ GamePlayer) error {
		g.startGame()
		g.messageAllPlayers("We are ready to play")
		return nil
	})

}

func (g *Game) startPickTimeout(player GamePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelPickTimeout = cancel
	g.startTimeout(ctx, pickTimeout, &Player{}, func(_ GamePlayer) error {
		catIdx, valIdx := g.firstAvailableQuestion()
		return g.processPick(player, catIdx, valIdx)
	})
}

func (g *Game) startBuzzTimeout(player GamePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	g.StartBuzzCountdown = true
	g.cancelBuzzTimeout = cancel
	g.startTimeout(ctx, buzzTimeout, &Player{}, func(_ GamePlayer) error {
		g.skipQuestion()
		return nil
	})
}

func (g *Game) startAnswerTimeout(player GamePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	player.setCancelAnswerTimeout(cancel)
	answerTimeout := defaultAnsTimeout
	if g.CurQuestion.DailyDouble {
		answerTimeout = dailyDoubleAnsTimeout
	} else if g.Round == FinalRound {
		answerTimeout = finalJeopardyAnsTimeout
		g.StartFinalAnswerCountdown = true
	}
	go g.startTimeout(ctx, answerTimeout, player, func(player GamePlayer) error {
		if g.Round == FinalRound {
			return g.processFinalRoundAns(player, false, "answer-timeout")
		}
		g.nextQuestion(player, false)
		return nil
	})
}

func (g *Game) startVoteTimeout(player GamePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelVoteTimeout = cancel
	g.startTimeout(ctx, voteTimeout, &Player{}, func(_ GamePlayer) error {
		g.nextQuestion(g.LastToAnswer, g.AnsCorrectness)
		return nil
	})
}

func (g *Game) startWagerTimeout(player GamePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	player.setCancelWagerTimeout(cancel)
	wagerTimeout := dailyDoubleWagerTimeout
	if g.Round == FinalRound {
		wagerTimeout = finalJeopardyWagerTimeout
		g.StartFinalWagerCountdown = true
	}
	g.startTimeout(ctx, wagerTimeout, player, func(player GamePlayer) error {
		wager := 5
		if g.Round == FinalRound {
			wager = 0
		}
		return g.processWager(player, wager)
	})
}
