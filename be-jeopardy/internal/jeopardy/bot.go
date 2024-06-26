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

const (
	botPickTimeout    = 3 * time.Second
	botPassTimeout    = 5 * time.Second
	botBuzzTimeout    = 20 * time.Second
	botAnswerTimeout  = 5 * time.Second
	botDDAnsTimeout   = 10 * time.Second
	botWagerTimeout   = 5 * time.Second
	botDisputeTimeout = 10 * time.Second
)

var botConfigs = []struct {
	name   string
	imgUrl string
}{
	{"Eager Eagle", "https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/eagle.png"},
	{"Golden Gorilla", "https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/gorilla.png"},
	{"Sharp Shark", "https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/shark.png"},
	{"Smart Snake", "https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/snake.png"},
	{"Tough Tiger", "https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/tiger.png"},
}

func NewBot(name string, i int) *Bot {
	botConfig := botConfigs[i%len(botConfigs)]
	bot := &Bot{
		Player:  NewPlayer(botConfig.name, botConfig.imgUrl, ""),
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
		for msg := range p.botChan {
			cancel()
			ctx, cancel = context.WithCancel(context.Background())
			go p.processMessage(ctx, msg)
		}
	}()
}

func sendBuzzAfter(ctx context.Context, g *Game, msg Message, passDelay, buzzDelay time.Duration) {
	ticker := time.NewTicker(1 * time.Second)
	secondsSinceHumansPassed := 0
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
			humanPasses := 0
			for _, player := range g.Players {
				if !player.isBot() && !player.canBuzz() {
					humanPasses++
				}
			}
			if humanPasses == g.numHumans() {
				secondsSinceHumansPassed++
			}
			if secondsSinceHumansPassed > 3 {
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

func (p *Bot) processMessage(ctx context.Context, resp Response) {
	g := resp.Game
	if g.Paused {
		return
	}
	msg := Message{
		Player: p,
		State:  g.State,
	}
	switch g.State {
	case RecvPick:
		if !p.canPick() {
			return
		}
		msg.CatIdx, msg.ValIdx = g.nextQuestionInCategory()
		wrongAnswer := false
		for _, ans := range g.CurQuestion.Answers {
			if !ans.Correct {
				wrongAnswer = true
				break
			}
		}
		timeout := botPickTimeout
		if wrongAnswer {
			timeout = 10 * time.Second
		}
		timeout = min(timeout, time.Duration(g.PickTimeout-1)*time.Second)
		sendMessageAfter(ctx, g, msg, timeout)
	case RecvBuzz:
		if !p.canBuzz() {
			return
		}
		scores := sortScores(g.Players)
		msg.IsPass = p.score() != scores[len(scores)-1]
		buzzTimeout := min(botBuzzTimeout, time.Duration(g.BuzzTimeout-1)*time.Second)
		sendBuzzAfter(ctx, g, msg, botPassTimeout, buzzTimeout)
	case RecvAns:
		if !p.canAnswer() {
			return
		}
		msg.Answer = g.CurQuestion.Answer
		timeout := botAnswerTimeout
		if g.CurQuestion.DailyDouble {
			timeout = botDDAnsTimeout
		}
		timeout = min(timeout, time.Duration(g.AnswerTimeout-1)*time.Second)
		sendMessageAfter(ctx, g, msg, timeout)
	case RecvWager:
		if !p.canWager() {
			return
		}
		msg.Wager = p.pickWager(g.Players, g.roundMax())
		timeout := min(botWagerTimeout, time.Duration(g.WagerTimeout-1)*time.Second)
		sendMessageAfter(ctx, g, msg, timeout)
	case RecvDispute:
		if !p.canDispute() {
			return
		}
		msg.Dispute = true
		sendMessageAfter(ctx, g, msg, botDisputeTimeout)
	case PostGame:
		p.setPlayAgain(true)
	case PreGame, BoardIntro:
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
	scores := []int{}
	for _, player := range players {
		scores = append(scores, player.score())
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i] > scores[j]
	})
	return scores
}

func (p *Bot) copyState(player GamePlayer) {
	p.Score = player.score()
	p.CanPick = player.canPick()
	p.CanBuzz = player.canBuzz()
	p.CanAnswer = player.canAnswer()
	p.CanWager = player.canWager()
	p.CanDispute = player.canDispute()
	p.FinalWager = player.finalWager()
	p.FinalCorrect = player.finalCorrect()
	p.FinalProtestors = player.finalProtestors()
	p.PlayAgain = player.playAgain()
}

func (p *Bot) isBot() bool {
	return true
}
