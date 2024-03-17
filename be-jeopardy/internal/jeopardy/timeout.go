package jeopardy

import (
	"context"
	"time"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
)

const boardIntroTimeout = 25

type GameTimeouts struct {
	cancelBoardIntroTimeout context.CancelFunc
	cancelPickTimeout       context.CancelFunc
	cancelBuzzTimeout       context.CancelFunc
	cancelVoteTimeout       context.CancelFunc
	cancelDisputeTimeout    context.CancelFunc
}

func (g *Game) startTimeout(ctx context.Context, timeout int, player GamePlayer, processTimeout func(player GamePlayer) error) {
	go func() {
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
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

func (g *Game) startBoardIntroTimeout() {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelBoardIntroTimeout = cancel
	g.startTimeout(ctx, boardIntroTimeout, &Player{}, func(_ GamePlayer) error {
		if g.Round == FirstRound {
			g.startGame()
		} else {
			g.setState(RecvPick, g.lowestPlayer())
		}
		g.messageAllPlayers("We are ready to play")
		return nil
	})

}

func (g *Game) startPickTimeout(player GamePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelPickTimeout = cancel
	g.startTimeout(ctx, g.PickTimeout, &Player{}, func(_ GamePlayer) error {
		catIdx, valIdx := g.firstAvailableQuestion()
		return g.processPick(player, catIdx, valIdx)
	})
}

func (g *Game) startBuzzTimeout() {
	ctx, cancel := context.WithCancel(context.Background())
	g.StartBuzzCountdown = true
	g.cancelBuzzTimeout = cancel
	g.startTimeout(ctx, g.BuzzTimeout, &Player{}, func(_ GamePlayer) error {
		g.skipQuestion()
		return nil
	})
}

func (g *Game) startAnswerTimeout(player GamePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	player.setCancelAnswerTimeout(cancel)
	timeout := g.AnswerTimeout
	if g.Round == FinalRound {
		timeout = g.FinalAnswerTimeout
		g.StartFinalAnswerCountdown = true
	}
	go g.startTimeout(ctx, timeout, player, func(player GamePlayer) error {
		if g.Round == FinalRound {
			return g.processFinalRoundAns(player, false, "answer-timeout")
		}
		g.CurQuestion.CurAns = &Answer{
			Player:  player,
			Answer:  "answer-timeout",
			Correct: false,
			Bot:     player.isBot(),
		}
		g.CurQuestion.Answers = append(g.CurQuestion.Answers, g.CurQuestion.CurAns)
		g.nextQuestion(player, false)
		return nil
	})
}

func (g *Game) startVoteTimeout() {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelVoteTimeout = cancel
	g.startTimeout(ctx, g.VoteTimeout, &Player{}, func(_ GamePlayer) error {
		g.nextQuestion(g.CurQuestion.CurAns.Player, g.AnsCorrectness)
		return nil
	})
}

func (g *Game) startDisputeTimeout() {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancelDisputeTimeout = cancel
	g.startTimeout(ctx, g.DisputeTimeout, &Player{}, func(_ GamePlayer) error {
		g.Disputers = 0
		g.NonDisputers = 0
		g.setState(RecvPick, g.DisputePicker)
		g.messageAllPlayers("Dispute resolved")
		return nil
	})
}

func (g *Game) startWagerTimeout(player GamePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	player.setCancelWagerTimeout(cancel)
	wagerTimeout := g.WagerTimeout
	if g.Round == FinalRound {
		wagerTimeout = g.FinalWagerTimeout
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
