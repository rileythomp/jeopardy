package jeopardy

import (
	"context"
	"time"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/socket"
)

type Bot struct {
	*Player
	botChan chan Response
}

func NewBot(name string) *Bot {
	bot := &Bot{
		Player:  NewPlayer(name),
		botChan: make(chan Response),
	}
	bot.Conn = socket.NewSafeConn(nil) // so bot is treated as connected by front-end
	return bot
}

func (p *Bot) sendMessage(msg Response) error {
	p.botChan <- msg
	return nil
}

func (p *Bot) processMessages() {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		for {
			select {
			case msg := <-p.botChan:
				cancel()
				ctx, cancel = context.WithCancel(context.Background())
				go p.processMessage(ctx, msg)
			}
		}
	}()
}

// TODO: Improve bot logic
func (p *Bot) processMessage(ctx context.Context, resp Response) {
	g := resp.Game
	if g.Paused {
		return
	}

	msg := Message{Player: p}
	switch g.State {
	case RecvPick:
		if !p.canPick() {
			return
		}
		msg.CatIdx, msg.ValIdx = g.firstAvailableQuestion()
	case RecvBuzz:
		if !p.canBuzz() {
			return
		}
		msg.IsPass = false
	case RecvAns:
		if !p.canAnswer() {
			return
		}
		msg.Answer = g.CurQuestion.Answer
	case RecvVote:
		if !p.canVote() {
			return
		}
		msg.Confirm = true
	case RecvWager:
		if !p.canWager() {
			return
		}
		msg.Wager = 10
	case PreGame, PostGame:
		return
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(5 * time.Second):
		g.msgChan <- msg
	}

}
