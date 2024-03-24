package jeopardy

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
)

type ChatMessage struct {
	PlayerName string `json:"name"`
	Message    string `json:"message"`
	TimeStamp  int64  `json:"timeStamp"`
}

func JoinGameChat(playerId string, conn SafeConn) error {
	game, err := GetPlayerGame(playerId)
	if err != nil {
		return err
	}

	player, err := game.getPlayerById(playerId)
	if err != nil {
		return err
	}
	if player.chatConn() != nil {
		return fmt.Errorf("Player already in chat")
	}
	player.setChatConn(conn)

	player.sendChatPings()
	player.processChatMessages(game.chatChan)

	return nil
}

func (p *Player) processChatMessages(chatChan chan ChatMessage) {
	go func() {
		log.Infof("Starting to process chat messages for player %s", p.Name)
		for {
			message, err := p.readChatMessage()
			if err != nil {
				log.Errorf("Error reading chat message from player %s: %s", p.Name, err.Error())
				if websocket.IsCloseError(err, 1001) {
					log.Infof("Player %s closed chat connection", p.Name)
				}
				return
			}
			var msg ChatMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Errorf("Error parsing chat message: %s", err.Error())
			}
			msg.PlayerName = p.Name
			msg.TimeStamp = time.Now().Unix()
			chatChan <- msg
		}
	}()
}

func (g *Game) processChatMessages() {
	go func() {
		for {
			select {
			case msg := <-g.chatChan:
				for _, p := range g.Players {
					_ = p.sendChatMessage(msg)
				}
			}
		}
	}()
}

func (p *Player) readChatMessage() ([]byte, error) {
	if p.ChatConn == nil {
		log.Infof("Skipping reading chat message from player %s because connection is nil", p.Name)
		return nil, fmt.Errorf("Player %s has no chat connection", p.Name)
	}
	_, msg, err := p.ChatConn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (p *Player) sendChatMessage(msg ChatMessage) error {
	if p.ChatConn == nil {
		log.Errorf("Error sending chat message to player %s because connection is nil", p.Name)
		return fmt.Errorf("player has no chat connection")
	}
	if err := p.ChatConn.WriteJSON(msg); err != nil {
		log.Errorf("Error sending chat message to player %s: %s", p.Name, err.Error())
		return fmt.Errorf("error sending chat message to player")
	}
	return nil
}

func (p *Player) sendChatPings() {
	go func() {
		log.Infof("Starting to send chat pings to player %s", p.Name)
		pingErrors := 0
		for {
			select {
			case <-p.sendChatPing.C:
				if err := p.sendChatMessage(ChatMessage{
					PlayerName: ping,
					Message:    ping,
					TimeStamp:  time.Now().Unix(),
				}); err != nil {
					if p.ChatConn == nil {
						log.Infof("Stopping sending chat pings to player %s because connection is nil", p.Name)
						return
					}
					pingErrors++
					if pingErrors >= 3 {
						log.Infof("Too many chat ping errors, closing connection to player %s", p.Name)
						if err := p.ChatConn.Close(); err != nil {
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
