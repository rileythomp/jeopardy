package jeopardy

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/socket"

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
	CanVote         bool            `json:"canVote"`
	FinalWager      int             `json:"finalWager"`
	FinalAnswer     string          `json:"finalAnswer"`
	FinalCorrect    bool            `json:"finalCorrect"`
	FinalProtestors map[string]bool `json:"finalProtestors"`
	PlayAgain       bool            `json:"playAgain"`

	Conn     SafeConn `json:"conn"`
	ChatConn SafeConn `json:"chatConn"`

	cancelAnswerTimeout context.CancelFunc
	cancelWagerTimeout  context.CancelFunc

	sendGamePing *time.Ticker
	sendChatPing *time.Ticker
}

const (
	pingFrequency = 50 * time.Second
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
		CanVote:             false,
		FinalProtestors:     map[string]bool{},
		cancelAnswerTimeout: func() {},
		cancelWagerTimeout:  func() {},
		sendGamePing:        time.NewTicker(pingFrequency),
		sendChatPing:        time.NewTicker(pingFrequency),
	}
}

func (p *Player) processMessages(msgChan chan Message, pauseChan chan *Player) {
	go func() {
		log.Infof("Starting to process messages for player %s", p.Name)
		for {
			message, err := p.readMessage()
			if err != nil {
				log.Errorf("Error reading message from player %s: %s", p.Name, err.Error())
				if websocket.IsCloseError(err, 1001) {
					log.Infof("Player %s closed connection", p.Name)
				}
				pauseChan <- p
				return
			}
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Errorf("Error parsing message: %s", err.Error())
			}
			msg.Player = p
			msgChan <- msg
		}
	}()
}

func (p *Player) sendPings() {
	go func() {
		log.Infof("Starting to send pings to player %s", p.Name)
		pingErrors := 0
		for {
			select {
			case <-p.sendGamePing.C:
				if err := p.sendMessage(Response{
					Code:    socket.Info,
					Message: ping,
				}); err != nil {
					if p.Conn == nil {
						log.Infof("Stopping sending pings to player %s because connection is nil", p.Name)
						return
					}
					pingErrors++
					if pingErrors >= 3 {
						log.Infof("Too many ping errors, closing connection to player %s", p.Name)
						if err := p.Conn.Close(); err != nil {
							log.Errorf("Error closing connection: %s", err.Error())
						}
						return
					}
					continue
				}
				pingErrors = 0
			}
		}
	}()
}

func (p *Player) pausePlayer() {
	p.cancelAnswerTimeout()
	p.cancelWagerTimeout()
}

func (p *Player) resetPlayer() {
	p.Score = 0
	p.updateActions(false, false, false, false, false)
	p.FinalWager = 0
	p.FinalAnswer = ""
	p.FinalCorrect = false
	p.FinalProtestors = map[string]bool{}
	p.PlayAgain = false
}

func (p *Player) updateActions(pick, buzz, answer, wager, vote bool) {
	p.CanPick = pick
	p.CanBuzz = buzz
	p.CanAnswer = answer
	p.CanWager = wager
	p.CanVote = vote
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

func (p *Player) canBuzz(guessedWrong, passed []string) bool {
	return !p.inLists(guessedWrong, passed)
}

func (p *Player) canVote(confirmers, challengers []string) bool {
	return !p.inLists(confirmers, challengers)
}

func (p *Player) inLists(lists ...[]string) bool {
	for _, list := range lists {
		for _, id := range list {
			if id == p.Id {
				return true
			}
		}
	}
	return false
}

func (p *Player) readMessage() ([]byte, error) {
	if p.Conn == nil {
		log.Infof("Skipping reading message from player %s because connection is nil", p.Name)
		return nil, fmt.Errorf("Player %s has no connection", p.Name)
	}
	_, msg, err := p.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (p *Player) sendMessage(msg Response) error {
	if p.Conn == nil {
		log.Errorf("Error sending message to player %s because connection is nil", p.Name)
		return fmt.Errorf("player has no connection")
	}
	if err := p.Conn.WriteJSON(msg); err != nil {
		log.Errorf("Error sending message to player %s: %s", p.Name, err.Error())
		return fmt.Errorf("error sending message to player")
	}
	return nil
}
