package jeopardy

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
)

type Reaction struct {
	PlayerName string `json:"name"`
	Reaction   string `json:"reaction"`
	TimeStamp  int64  `json:"timeStamp"`
	RandPos    int    `json:"randPos"`
}

func JoinReactions(playerId string, conn SafeConn) error {
	game, err := GetPlayerGame(playerId)
	if err != nil {
		return err
	}

	player, err := game.getPlayerById(playerId)
	if err != nil {
		return err
	}
	if player.reactionConn() != nil {
		return fmt.Errorf("Player already in game reactions")
	}
	player.setReactionConn(conn)

	player.sendReactionPings()
	player.processReactions(game.reactChan)

	return nil
}

func (p *Player) processReactions(reactChan chan Reaction) {
	go func() {
		log.Infof("Starting to process reaction messages for player %s", p.Name)
		for {
			message, err := p.readReaction()
			if err != nil {
				log.Errorf("Error reading reaction message from player %s: %s", p.Name, err.Error())
				if websocket.IsCloseError(err, 1001) {
					log.Infof("Player %s closed reaction connection", p.Name)
				}
				return
			}
			var msg Reaction
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Errorf("Error parsing reaction message: %s", err.Error())
			}
			msg.PlayerName = p.Name
			msg.TimeStamp = time.Now().Unix()
			msg.RandPos = getRandPos(10, 150)
			reactChan <- msg
		}
	}()
}

func (p *Player) readReaction() ([]byte, error) {
	if p.ReactionConn == nil {
		log.Infof("Skipping reading reaction from player %s because connection is nil", p.Name)
		return nil, fmt.Errorf("Player %s has no reaction connection", p.Name)
	}
	_, msg, err := p.ReactionConn.ReadMessage()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (p *Player) sendReaction(msg Reaction) error {
	if p.ReactionConn == nil {
		return fmt.Errorf("player has no reaction connection")
	}
	if err := p.ReactionConn.WriteJSON(msg); err != nil {
		log.Errorf("Error sending reaction to player %s: %s", p.Name, err.Error())
		return fmt.Errorf("error sending reaction to player")
	}
	return nil
}

func (p *Player) sendReactionPings() {
	go func() {
		log.Infof("Starting to send reaction pings to player %s", p.Name)
		pingErrors := 0
		for {
			select {
			case <-p.sendReactPing.C:
				if err := p.sendReaction(Reaction{
					PlayerName: ping,
					Reaction:   ping,
					TimeStamp:  time.Now().Unix(),
					RandPos:    getRandPos(10, 10),
				}); err != nil {
					if p.ReactionConn == nil {
						log.Infof("Stopping sending reaction pings to player %s because connection is nil", p.Name)
						return
					}
					pingErrors++
					if pingErrors >= 3 {
						log.Infof("Too many reaction ping errors, closing connection to player %s", p.Name)
						if err := p.ReactionConn.Close(); err != nil {
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

func getRandPos(min, max int) int {
	return rand.Intn(max-min+1) + min
}
