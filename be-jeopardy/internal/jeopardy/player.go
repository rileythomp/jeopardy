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

type GamePlayer interface {
	id() string
	name() string
	email() string
	conn() SafeConn
	chatConn() SafeConn
	reactionConn() SafeConn
	score() int
	canPick() bool
	canBuzz() bool
	canAnswer() bool
	canWager() bool
	canDispute() bool
	finalWager() int
	finalCorrect() bool
	finalProtestors() map[string]bool
	playAgain() bool
	isBot() bool

	setId(string)
	setName(string)
	setImg(string)
	setConn(SafeConn)
	setChatConn(SafeConn)
	setReactionConn(SafeConn)
	setCanBuzz(bool)
	setCanAnswer(bool)
	setCanWager(bool)
	setCanDispute(bool)
	setFinalWager(int)
	setFinalAnswer(string)
	setFinalCorrect(bool)
	setPlayAgain(bool)

	readMessages(msgChan chan Message, disconnectChan chan GamePlayer)
	processChatMessages(chan ChatMessage)
	processReactions(chan Reaction)
	sendPings()
	sendChatPings()
	sendReactionPings()

	sendMessage(Response) error
	sendChatMessage(ChatMessage) error
	sendReaction(Reaction) error
	updateActions(pick, buzz, answer, wager bool)
	updateScore(val int, isCorrect, penalty bool, round RoundState)
	addFinalProtestor(string)
	addToScore(int)
	resetPlayer()
	pausePlayer()
	endConnections()

	setCancelAnswerTimeout(context.CancelFunc)
	setCancelWagerTimeout(context.CancelFunc)
	cancelAnswerTimeout()
	cancelWagerTimeout()
}

type Player struct {
	Id              string          `json:"id"`
	Name            string          `json:"name"`
	Email           string          `json:"email"`
	Score           int             `json:"score"`
	CanPick         bool            `json:"canPick"`
	CanBuzz         bool            `json:"canBuzz"`
	CanAnswer       bool            `json:"canAnswer"`
	CanWager        bool            `json:"canWager"`
	CanDispute      bool            `json:"canDispute"`
	FinalWager      int             `json:"finalWager"`
	FinalAnswer     string          `json:"finalAnswer"`
	FinalCorrect    bool            `json:"finalCorrect"`
	FinalProtestors map[string]bool `json:"finalProtestors"`
	PlayAgain       bool            `json:"playAgain"`
	ImgUrl          string          `json:"imgUrl"`

	Conn         SafeConn `json:"conn"`
	ChatConn     SafeConn `json:"chatConn"`
	ReactionConn SafeConn `json:"reactionConn"`

	CancelAnswerTimeout context.CancelFunc `json:"-"`
	CancelWagerTimeout  context.CancelFunc `json:"-"`

	sendGamePing  *time.Ticker
	sendChatPing  *time.Ticker
	sendReactPing *time.Ticker
}

const (
	pingFrequency = 50 * time.Second
	ping          = "ping"
)

var playerImgs = []string{
	"https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/cat.png",
	"https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/deer.png",
	"https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/dragon.png",
	"https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/giraffe.png",
	"https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/panda.png",
	"https://xdlhyjzjygansfeoguvs.supabase.co/storage/v1/object/public/jeopardy_imgs/lion.png",
}

func NewPlayer(name, imgUrl, email string) *Player {
	return &Player{
		Id:                  uuid.New().String(),
		Name:                name,
		Email:               email,
		Score:               0,
		CanPick:             false,
		CanBuzz:             false,
		CanAnswer:           false,
		CanWager:            false,
		FinalProtestors:     map[string]bool{},
		ImgUrl:              imgUrl,
		CancelAnswerTimeout: func() {},
		CancelWagerTimeout:  func() {},
		sendGamePing:        time.NewTicker(pingFrequency),
		sendChatPing:        time.NewTicker(pingFrequency),
	}
}

func (p *Player) readMessages(msgChan chan Message, disconnectChan chan GamePlayer) {
	go func() {
		log.Infof("Starting to read messages from player %s", p.Name)
		for {
			message, err := p.readMessage()
			if err != nil {
				log.Errorf("Error reading message from player %s: %s", p.Name, err.Error())
				if websocket.IsCloseError(err, 1001) {
					log.Infof("Player %s closed connection", p.Name)
				}
				disconnectChan <- p
				return
			}
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Errorf("Error parsing message from player: %s", err.Error())
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
	p.CancelAnswerTimeout()
	p.CancelWagerTimeout()
}

func (p *Player) resetPlayer() {
	p.Score = 0
	p.updateActions(false, false, false, false)
	p.FinalWager = 0
	p.FinalAnswer = ""
	p.FinalCorrect = false
	p.FinalProtestors = map[string]bool{}
	p.PlayAgain = false
}

func (p *Player) updateActions(pick, buzz, answer, wager bool) {
	p.CanPick = pick
	p.CanBuzz = buzz
	p.CanAnswer = answer
	p.CanWager = wager
}

func (p *Player) updateScore(val int, isCorrect, penalty bool, round RoundState) {
	if round == FinalRound {
		val = p.FinalWager
	}
	if !isCorrect {
		val *= -1
		if !penalty {
			val = 0
		}
	}
	p.Score += val
}

func (p *Player) id() string {
	return p.Id
}

func (p *Player) name() string {
	return p.Name
}

func (p *Player) email() string {
	return p.Email
}

func (p *Player) conn() SafeConn {
	return p.Conn
}

func (p *Player) chatConn() SafeConn {
	return p.ChatConn
}

func (p *Player) reactionConn() SafeConn {
	return p.ReactionConn
}

func (p *Player) score() int {
	return p.Score
}

func (p *Player) canPick() bool {
	return p.CanPick
}

func (p *Player) canBuzz() bool {
	return p.CanBuzz
}

func (p *Player) canAnswer() bool {
	return p.CanAnswer
}

func (p *Player) canWager() bool {
	return p.CanWager
}

func (p *Player) canDispute() bool {
	return p.CanDispute
}

func (p *Player) finalWager() int {
	return p.FinalWager
}

func (p *Player) finalCorrect() bool {
	return p.FinalCorrect
}

func (p *Player) finalProtestors() map[string]bool {
	return p.FinalProtestors
}

func (p *Player) playAgain() bool {
	return p.PlayAgain
}

func (p *Player) isBot() bool {
	return false
}

func (p *Player) setId(id string) {
	p.Id = id
}

func (p *Player) setName(name string) {
	p.Name = name
}

func (p *Player) setImg(img string) {
	p.ImgUrl = img
}

func (p *Player) setConn(conn SafeConn) {
	p.Conn = conn
}

func (p *Player) setChatConn(conn SafeConn) {
	p.ChatConn = conn
}

func (p *Player) setReactionConn(conn SafeConn) {
	p.ReactionConn = conn
}

func (p *Player) setCanBuzz(canBuzz bool) {
	p.CanBuzz = canBuzz
}

func (p *Player) setCanAnswer(canAnswer bool) {
	p.CanAnswer = canAnswer
}

func (p *Player) setCanWager(canWager bool) {
	p.CanWager = canWager
}

func (p *Player) setCanDispute(canDispute bool) {
	p.CanDispute = canDispute
}

func (p *Player) setFinalWager(wager int) {
	p.FinalWager = wager
}

func (p *Player) setFinalAnswer(answer string) {
	p.FinalAnswer = answer
}

func (p *Player) setFinalCorrect(correct bool) {
	p.FinalCorrect = correct
}

func (p *Player) setPlayAgain(playAgain bool) {
	p.PlayAgain = playAgain
}

func (p *Player) addFinalProtestor(playerId string) {
	p.FinalProtestors[playerId] = true
}

func (p *Player) addToScore(points int) {
	p.Score += points
}

func (p *Player) endConnections() {
	p.Conn = nil
	p.ChatConn = nil
}

func (p *Player) setCancelWagerTimeout(cancel context.CancelFunc) {
	p.CancelWagerTimeout = cancel
}

func (p *Player) setCancelAnswerTimeout(cancel context.CancelFunc) {
	p.CancelAnswerTimeout = cancel
}

func (p *Player) cancelAnswerTimeout() {
	p.CancelAnswerTimeout()
}

func (p *Player) cancelWagerTimeout() {
	p.CancelWagerTimeout()
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
