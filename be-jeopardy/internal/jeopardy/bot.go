package jeopardy

import (
	"context"
	"fmt"
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
	bot.Conn = socket.NewSafeConn(nil) // so bot is treated as connected by front-
	return bot
}

func (p *Bot) sendMessage(msg Response) error {
	p.botChan <- msg
	return nil
}

func (p *Bot) handleMessage(ctx context.Context, game *Game, f func()) {
	select {
	case <-ctx.Done():
		fmt.Printf("an action occurred in the game that is causing bot %s to stop waiting\n", p.name())
		break
	case <-time.After(5 * time.Second):
		f()
	}
}

func (p *Bot) processMessage(ctx context.Context, msg Response) {
	g := msg.Game
	if g.Paused {
		fmt.Printf("Bot %s says the game is paused\n", p.name())
		return
	}
	switch g.State {
	case RecvPick:
		fmt.Printf("Bot %s says it's time to pick\n", p.name())
		if !p.canPick() {
			break
		}
		fmt.Printf("Bot %s will wait a few seconds to pick\n", p.name())
		p.handleMessage(ctx, g, func() {
			fmt.Printf("Bot %s is done waiting to pick\n", p.name())
			c, v := g.firstAvailableQuestion()
			resp := Message{
				Player: p,
				PickMessage: PickMessage{
					CatIdx: c,
					ValIdx: v,
				},
			}
			fmt.Printf("Bot %s is picking category %d and value %d\n", p.name(), c, v)
			g.msgChan <- resp
		})
		// select {
		// case <-ctx.Done():
		// 	fmt.Printf("an action occurred in the game that is causing bot %s to stop picking\n", p.name())
		// 	break
		// case <-time.After(5 * time.Second):
		// 	fmt.Printf("Bot %s is done waiting to pick\n", p.name())
		// 	c, v := g.firstAvailableQuestion()
		// 	resp := Message{
		// 		Player: p,
		// 		PickMessage: PickMessage{
		// 			CatIdx: c,
		// 			ValIdx: v,
		// 		},
		// 	}
		// 	fmt.Printf("Bot %s is picking category %d and value %d\n", p.name(), c, v)
		// 	g.msgChan <- resp
		// }
	case RecvBuzz:
		fmt.Printf("Bot %s says it's time to buzz\n", p.name())
		if !p.canBuzz() {
			break
		}
		fmt.Printf("Bot %s will wait a few seconds to buzz\n", p.name())
		select {
		case <-ctx.Done():
			fmt.Printf("an action occurred in the game that is causing bot %s to stop buzzing\n", p.name())
			break
		case <-time.After(5 * time.Second):
			fmt.Printf("Bot %s is done waiting to buzz\n", p.name())
			resp := Message{
				Player: p,
				BuzzMessage: BuzzMessage{
					IsPass: false,
				},
			}
			fmt.Printf("Bot %s is answering\n", p.name())
			g.msgChan <- resp
		}
	case RecvAns:
		fmt.Printf("Bot %s says it's time to answer\n", p.name())
		if !p.canAnswer() {
			break
		}
		fmt.Printf("Bot %s will wait a few seconds to answer\n", p.name())
		select {
		case <-ctx.Done():
			fmt.Printf("an action occurred in the game that is causing bot %s to stop answering\n", p.name())
			break
		case <-time.After(5 * time.Second):
			fmt.Printf("Bot %s is done waiting to answer\n", p.name())
			resp := Message{
				Player: p,
				AnswerMessage: AnswerMessage{
					Answer: g.CurQuestion.Answer,
				},
			}
			fmt.Printf("Bot %s is answering %s\n", p.name(), g.CurQuestion.Answer)
			g.msgChan <- resp
		}
	case RecvVote:
		fmt.Printf("Bot %s says it's time to vote\n", p.name())
		if !p.canVote() {
			break
		}
		fmt.Printf("Bot %s will wait a few seconds to vote\n", p.name())
		select {
		case <-ctx.Done():
			fmt.Printf("an action occurred in the game that is causing bot %s to stop voting\n", p.name())
			break
		case <-time.After(5 * time.Second):
			fmt.Printf("Bot %s is done waiting to vote\n", p.name())
			resp := Message{
				Player: p,
				VoteMessage: VoteMessage{
					Confirm: true,
				},
			}
			fmt.Printf("Bot %s is voting to confirm\n", p.name())
			g.msgChan <- resp
		}
	case RecvWager:
		fmt.Printf("Bot %s says it's time to wager\n", p.name())
		if !p.canWager() {
			break
		}
		fmt.Printf("Bot %s will wait a few seconds to wager\n", p.name())
		select {
		case <-ctx.Done():
			fmt.Printf("an action occurred in the game that is causing bot %s to stop wagering\n", p.name())
			break
		case <-time.After(5 * time.Second):
			fmt.Printf("Bot %s is done waiting to wager\n", p.name())
			resp := Message{
				Player: p,
				WagerMessage: WagerMessage{
					Wager: 10,
				},
			}
			fmt.Printf("Bot %s is wagering 10\n", p.name())
			g.msgChan <- resp
		}
	case PostGame:
		fmt.Printf("Bot %s says it is post game\n", p.name())
	case PreGame:
		fmt.Printf("Bot %s says it is pre game\n", p.name())
	}
}
