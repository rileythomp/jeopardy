package jeopardy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

	Conn SafeConn `json:"conn"`

	cancelAnswerTimeout context.CancelFunc
	cancelWagerTimeout  context.CancelFunc

	stopSendingPings chan bool
	sendPingTicker   *time.Ticker
}

const (
	pingFrequency = 1 * time.Second
	ping          = "ping"
)

func NewPlayer(name string) *Player {
	return &Player{
		Id:                  uuid.New().String(),
		Name:                name,
		Score:               0,
		CanPick:             false,
		CanBuzz:             false,
		CanAnswer:           false,
		CanWager:            false,
		CanConfirmAns:       false,
		FinalProtestors:     map[string]bool{},
		sendPingTicker:      time.NewTicker(pingFrequency),
		stopSendingPings:    make(chan bool),
		cancelAnswerTimeout: func() {},
		cancelWagerTimeout:  func() {},
	}
}

func (p *Player) processMessages(game *Game) {
	go func() {
		// TODO: USE A CHANNEL TO WAIT ON A MESSAGE OR TO END THE GAME
		for {
			msg, err := p.readMessage()
			if err != nil {
				log.Printf("Error reading message from player %s: %s\n", p.Name, err.Error())
				if websocket.IsCloseError(err, 1001) {
					log.Printf("Player %s closed connection\n", p.Name)
				}
				break
			}
			if err := game.processMsg(p, msg); err != nil {
				log.Printf("Error processing message: %s\n", err.Error())
				panic("error processing message")
			}
		}

		game.stopGame(p)
	}()
}

func (p *Player) stopPlayer() {
	p.cancelAnswerTimeout()
	p.cancelWagerTimeout()
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

func (p *Player) readMessage() ([]byte, error) {
	_, msg, err := p.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (p *Player) sendMessage(message any) error {
	msg, ok := message.(Response)
	if ok {
		if msg.Message != ping {
			fmt.Println("Sending message to player", p.Name, ":", msg.Message)
		}
	}

	if err := p.Conn.WriteJSON(message); err != nil {
		log.Printf("Error sending message to player %s: %s\n", p.Name, err.Error())
		return fmt.Errorf("error sending message to player")
	}
	return nil
}

func (p *Player) closeConnection() error {
	if err := p.Conn.Close(); err != nil {
		log.Printf("Error closing connection: %s\n", err.Error())
		return fmt.Errorf("error closing connection")
	}
	return nil
}
