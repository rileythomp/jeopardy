package jeopardy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type SafeConn interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteJSON(v interface{}) error
	Close() error
}

type Player struct {
	Id              string          `json:"id"`
	Name            string          `json:"name"`
	Score           int             `json:"score"`
	CanPick         bool            `json:"canPick"`
	CanBuzz         bool            `json:"canBuzz"`
	CanAnswer       bool            `json:"canAnswer"`
	CanWager        bool            `json:"canWager"`
	CanConfirmAns   bool            `json:"canConfirmAns"`
	FinalWager      int             `json:"finalWager"`
	FinalAnswer     string          `json:"finalAnswer"`
	FinalCorrect    bool            `json:"finalCorrect"`
	FinalProtestors map[string]bool `json:"finalProtestors"`

	conn SafeConn

	cancelAnswerTimeout context.CancelFunc
	cancelWagerTimeout  context.CancelFunc

	stopProcessingMessages chan bool
	stopSendingPings       chan bool
	sendPingTicker         *time.Ticker
}

const (
	pingFrequency = 10 * time.Second
	ping          = "ping"
)

func NewPlayer(name string) *Player {
	return &Player{
		Id:                     uuid.New().String(),
		Name:                   name,
		Score:                  0,
		CanPick:                false,
		CanBuzz:                false,
		CanAnswer:              false,
		CanWager:               false,
		CanConfirmAns:          false,
		FinalProtestors:        map[string]bool{},
		sendPingTicker:         time.NewTicker(pingFrequency),
		stopSendingPings:       make(chan bool),
		stopProcessingMessages: make(chan bool),
		cancelAnswerTimeout:    func() {},
		cancelWagerTimeout:     func() {},
	}
}

func (p *Player) terminate() {
	p.stopSendingPings <- true
	p.stopProcessingMessages <- true
	p.cancelAnswerTimeout()
	p.cancelWagerTimeout()
	if err := p.closeConnection(); err != nil {
		log.Printf("Error closing connection: %s\n", err.Error())
	}
}

func (p *Player) processMessages(game *Game) {
	go func() {
		// TODO: USE A CHANNEL TO WAIT ON A MESSAGE OR TO END THE GAME
		for {
			select {
			case <-p.stopProcessingMessages:
				return
			case msg := <-p.readMessage():
				err := game.processMsg(p, msg)
				if err != nil {
					log.Printf("Error handling request: %s\n", err.Error())
					p.closeConnWithMsg(err.Error())
					return
				}
			}
		}
	}()
}

func (p *Player) sendPings() {
	go func() {
		for {
			select {
			case <-p.stopSendingPings:
				return
			case <-p.sendPingTicker.C:
				if err := p.sendMessage(Response{
					Code:    http.StatusOK,
					Message: ping,
				}); err != nil {
					log.Printf("Error sending ping: %s\n", err.Error())
					p.closeConnWithMsg("Error sending ping")
					return
				}
			}
		}
	}()
}

func (p *Player) updateActions(pick, buzz, answer, wager, confirm bool) {
	p.CanPick = pick
	p.CanBuzz = buzz
	p.CanAnswer = answer
	p.CanWager = wager
	p.CanConfirmAns = confirm
}

func (p *Player) updateScore(val int, isCorrect bool, round RoundState) {
	if round == FinalRound {
		val = p.FinalWager
	}
	if !isCorrect {
		val *= -1
	}
	p.Score += val
}

func (p *Player) canBuzz(guessedWrong []string) bool {
	for _, id := range guessedWrong {
		if id == p.Id {
			return false
		}
	}
	return true
}

func (p *Player) readMessage() <-chan []byte {
	ch := make(chan []byte)
	go func() {
		_, msg, err := p.conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from player: %s\n", err.Error())
			return
		}
		ch <- msg
	}()
	return ch
}

func (p *Player) sendMessage(message any) error {
	if err := p.conn.WriteJSON(message); err != nil {
		log.Printf("Error sending message to player: %s\n", err.Error())
		return fmt.Errorf("error sending message to player")
	}
	return nil
}

func (p *Player) closeConnection() error {
	if err := p.conn.Close(); err != nil {
		log.Printf("Error closing connection: %s\n", err.Error())
		return fmt.Errorf("error closing connection")
	}
	return nil
}

func (p *Player) closeConnWithMsg(msg string) {
	_ = p.sendMessage(Response{Message: msg, Code: http.StatusInternalServerError})
	_ = p.closeConnection()
}
