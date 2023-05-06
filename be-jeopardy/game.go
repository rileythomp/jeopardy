package main

import (
	"github.com/gorilla/websocket"
)

type (
	Player struct {
		Id    string `json:"id"`
		Name  string `json:"name"`
		Score int    `json:"score"`
		conn  *websocket.Conn
	}

	Game struct {
		Players map[string]*Player `json:"players"`
	}
)

func NewGame() *Game {
	return &Game{Players: map[string]*Player{}}
}

func (g *Game) numPlayersReady() int {
	playersReady := 0
	for i := range g.Players {
		if g.Players[i].conn != nil {
			playersReady++
		}
	}
	return playersReady
}

func (g *Game) messagePlayers(resp any) error {
	for _, player := range game.Players {
		if player.conn != nil {
			if err := player.conn.WriteJSON(resp); err != nil {
				return err
			}
		}
	}
	return nil
}
