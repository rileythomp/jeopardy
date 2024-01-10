package jeopardy

import (
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

	cancelSendingPing chan bool
	sendPingTicker    *time.Ticker
}

const (
	pingFrequency = 10 * time.Second
	ping          = "ping"
)

func NewPlayer(name string) *Player {
	return &Player{
		Id:                uuid.New().String(),
		Name:              name,
		Score:             0,
		CanPick:           false,
		CanBuzz:           false,
		CanAnswer:         false,
		CanWager:          false,
		CanConfirmAns:     false,
		FinalProtestors:   map[string]bool{},
		cancelSendingPing: make(chan bool),
		sendPingTicker:    time.NewTicker(pingFrequency),
	}
}

func (p *Player) processMessages(game *Game) {
	go func() {
		// TODO: USE A CHANNEL TO WAIT ON A MESSAGE OR TO END THE GAME
		for {
			msg, err := p.readMessage()
			if err != nil {
				p.closeConnWithMsg("Error reading message from WebSocket")
				return
			}
			err = game.processMsg(p, msg)
			if err != nil {
				log.Printf("Error handling request: %s\n", err.Error())
				p.closeConnWithMsg(err.Error())
				return
			}
		}
	}()
}

func (p *Player) sendPings() {
	go func() {
		for {
			select {
			case <-p.cancelSendingPing:
				return
			case t := <-p.sendPingTicker.C:
				if err := p.sendMessage(Response{
					Code:    http.StatusOK,
					Message: ping,
				}); err != nil {
					log.Printf("Error sending ping to client: %s\n", err.Error())
					p.closeConnWithMsg("Error sending ping to client")
					return
				}
				fmt.Printf("sent ping to player: %s at %v\n", p.Name, t)
			}
		}
	}()
}

func (p *Player) stopSendingPings() {
	p.cancelSendingPing <- true
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
	_, msg, err := p.conn.ReadMessage()
	if err != nil {
		log.Printf("Error reading message from WebSocket: %s\n", err.Error())
		return nil, fmt.Errorf("error reading message from player")
	}
	return msg, nil
}

func (p *Player) sendMessage(message any) error {
	if err := p.conn.WriteJSON(message); err != nil {
		log.Printf("Error writing message to WebSocket: %s\n", err.Error())
		return fmt.Errorf("error sending message to player")
	}
	return nil
}

func (p *Player) closeConnection() error {
	if err := p.conn.Close(); err != nil {
		log.Printf("Error closing WebSocket: %s\n", err.Error())
		return fmt.Errorf("error closing connection")
	}
	return nil
}

func (p *Player) closeConnWithMsg(msg string) {
	_ = p.sendMessage(Response{Message: msg, Code: http.StatusInternalServerError})
	_ = p.closeConnection()
}
