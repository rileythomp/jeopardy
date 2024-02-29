package jeopardy

import (
	"context"
	"sort"
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
	bot.Conn = socket.NewSafeConn(nil) // so bot is treated as connected by frontend
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

func sendBuzzAfter(ctx context.Context, g *Game, msg Message, passDelay, buzzDelay time.Duration) {
	ticker := time.NewTicker(1 * time.Second)
	passedTicks := 0
	passDelayTimeout := time.After(passDelay)
	buzzDelayTimeout := time.After(buzzDelay)
	for {
		select {
		case <-ctx.Done():
			return
		case <-passDelayTimeout:
			if msg.IsPass {
				g.msgChan <- msg
				return
			}
		case <-buzzDelayTimeout:
			g.msgChan <- msg
			return
		case <-ticker.C:
			passes := 0
			for _, player := range g.Players {
				if !player.canBuzz() {
					passes++
				}
			}
			if passes > 1 {
				passedTicks++
			}
			if passedTicks > 3 {
				g.msgChan <- msg
				return
			}
		}
	}
}

func sendMessageAfter(ctx context.Context, g *Game, msg Message, delay time.Duration) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(delay):
		g.msgChan <- msg
	}
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
		msg.CatIdx, msg.ValIdx = g.nextQuestionInCategory()
		sendMessageAfter(ctx, g, msg, 5*time.Second)
	case RecvBuzz:
		if !p.canBuzz() {
			return
		}
		scores := sortScores(g.Players)
		msg.IsPass = p.score() != scores[2]
		sendBuzzAfter(ctx, g, msg, 5*time.Second, 20*time.Second)
	case RecvAns:
		if !p.canAnswer() {
			return
		}
		msg.Answer = g.CurQuestion.Answer
		delay := 5 * time.Second
		if g.CurQuestion.DailyDouble {
			delay = 10 * time.Second
		}
		sendMessageAfter(ctx, g, msg, delay)
	case RecvVote:
		if !p.canVote() {
			return
		}
		msg.Confirm = true
		sendMessageAfter(ctx, g, msg, 5*time.Second)
	case RecvWager:
		if !p.canWager() {
			return
		}
		msg.Wager = p.pickWager(g.Players, g.roundMax())
		sendMessageAfter(ctx, g, msg, 5*time.Second)
	case PreGame, PostGame:
		return
	}
}

func (p *Bot) pickWager(players []GamePlayer, roundMax int) int {
	scores := sortScores(players)
	if p.score() == scores[0] {
		return max(p.score()-max(scores[1], 0), roundMax)
	}
	return max(min(scores[0]-p.score(), p.score()), roundMax)
}

func sortScores(players []GamePlayer) []int {
	scores := []int{players[0].score(), players[1].score(), players[2].score()}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i] > scores[j]
	})
	return scores
}
