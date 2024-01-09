package jeopardy

import (
	"fmt"
	"log"
	"net/http"

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
}

func NewPlayer(name string) *Player {
	return &Player{
		Id:              uuid.New().String(),
		Name:            name,
		Score:           0,
		CanPick:         false,
		CanBuzz:         false,
		CanAnswer:       false,
		CanWager:        false,
		CanConfirmAns:   false,
		FinalProtestors: map[string]bool{},
	}
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

func (g *Game) addPlayer(name string) (string, error) {
	if g.State != PreGame {
		return "", fmt.Errorf("game already in progress")
	}
	if len(g.Players) > 2 {
		return "", fmt.Errorf("game is full")
	}
	player := NewPlayer(name)
	g.Players = append(g.Players, player)
	return player.Id, nil
}
