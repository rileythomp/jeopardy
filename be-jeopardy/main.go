package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type (
	Request struct {
		Token string `json:"token,omitempty"`
	}

	JoinRequest struct {
		Request
		PlayerName string `json:"playerName"`
	}

	PlayRequest struct {
		Request
	}

	PickRequest struct {
		Request
		TopicIdx int `json:"topicIdx"`
		ValIdx   int `json:"valIdx"`
	}

	BuzzRequest struct {
		Request
		IsPass bool `json:"isPass"`
	}

	AnswerRequest struct {
		Request
		Answer string `json:"answer"`
	}

	WagerRequest struct {
		Request
		Wager int `json:"wager"`
	}

	Response struct {
		Code      int     `json:"code"`
		Token     string  `json:"token,omitempty"`
		Message   string  `json:"message"`
		Game      *Game   `json:"game,omitempty"`
		CurPlayer *Player `json:"curPlayer,omitempty"`
	}
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == "http://localhost:4200"
		},
	}

	game = NewGame()
)

func closeConnWithMsg(conn *websocket.Conn, msg string, code int) {
	_ = conn.WriteJSON(Response{Message: msg, Code: code})
	conn.Close()
}

func joinGame(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade connection to WebSocket:", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Failed to read message from WebSocket:", err)
		closeConnWithMsg(conn, "Failed to read message from WebSocket", http.StatusInternalServerError)
		return
	}
	var joinReq JoinRequest
	if err := json.Unmarshal(msg, &joinReq); err != nil {
		log.Println("Failed to parse join request:", err)
		closeConnWithMsg(conn, "Failed to parse join request", http.StatusBadRequest)
		return
	}

	if game.hasStarted() {
		log.Println("Game already in progress")
		closeConnWithMsg(conn, "Game already in progress", http.StatusForbidden)
		return
	}
	playerId := game.addPlayer(joinReq.PlayerName)

	jwt, err := generateJWT(playerId)
	if err != nil {
		log.Println("Failed to generate token:", err)
		closeConnWithMsg(conn, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	resp := Response{
		Code:    200,
		Token:   jwt,
		Message: "Authorized to join game",
		Game:    game,
	}
	_ = conn.WriteJSON(resp)
	if err != nil {
		log.Println("Failed to write message to WebSocket:", err)
		conn.Close()
		return
	}
}

func playGame(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade connection to WebSocket:", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Failed to read message from WebSocket:", err)
		closeConnWithMsg(conn, "Failed to read message WebSocket", http.StatusInternalServerError)
		return
	}
	var playReq PlayRequest
	if err := json.Unmarshal(msg, &playReq); err != nil {
		log.Println("Failed to unmarshal play request:", err)
		closeConnWithMsg(conn, "Failed to parse join request", http.StatusBadRequest)
		return
	}

	playerId, err := getJWTSubject(playReq.Token)
	if err != nil {
		log.Println("Failed to get playerId from token:", err)
		closeConnWithMsg(conn, "Failed to get playerId from token", http.StatusForbidden)
		return
	}

	player := game.getPlayerById(playerId)
	if player == nil {
		log.Println("Player not found")
		closeConnWithMsg(conn, "Player not found", http.StatusForbidden)
		return
	}
	player.conn = conn

	resp := Response{
		Code:    200,
		Message: "Waiting for more players",
		Game:    game,
	}
	if game.readyToPlay() {
		if err := game.setQuestions(); err != nil {
			log.Println("Failed to set questions:", err)
			closeConnWithMsg(conn, "Failed to set questions", http.StatusInternalServerError)
			return
		}
		game.setState(RecvPick, game.Players[0].Id)
		resp = Response{
			Code:    200,
			Message: "We are ready to play",
			Game:    game,
		}
	}

	if err := game.messageAllPlayers(resp); err != nil {
		log.Println("Error sending message to players:", err)
		closeConnWithMsg(conn, "Error sending message to players", http.StatusInternalServerError)
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read message from WebSocket:", err)
			closeConnWithMsg(conn, "Failed to read message from WebSocket", http.StatusInternalServerError)
			return
		}

		switch game.State {
		case RecvPick:
			var pickReq PickRequest
			if err := json.Unmarshal(msg, &pickReq); err != nil {
				log.Println("Failed to parse pick request:", err)
				closeConnWithMsg(conn, "Failed to parse pick request", http.StatusBadRequest)
				return
			}
			playerId, err := getJWTSubject(pickReq.Token)
			if err != nil {
				log.Println("Failed to get playerId from token:", err)
				closeConnWithMsg(conn, "Failed to get playerId from token", http.StatusForbidden)
				return
			}
			err = game.handlePick(playerId, pickReq.TopicIdx, pickReq.ValIdx)
			if err != nil {
				log.Println("Failed to handle pick:", err)
				closeConnWithMsg(conn, fmt.Sprintf("Failed to handle pick: %s", err), http.StatusInternalServerError)
				return
			}
		case RecvBuzz:
			var buzzReq BuzzRequest
			if err := json.Unmarshal(msg, &buzzReq); err != nil {
				log.Println("Failed to parse buzz request:", err)
				closeConnWithMsg(conn, "Failed to parse buzz request", http.StatusBadRequest)
				return
			}
			playerId, err := getJWTSubject(buzzReq.Token)
			if err != nil {
				log.Println("Failed to get playerId from token:", err)
				closeConnWithMsg(conn, "Failed to get playerId from token", http.StatusForbidden)
				return
			}
			err = game.handleBuzz(playerId, buzzReq.IsPass)
			if err != nil {
				log.Println("Failed to handle buzz:", err)
				closeConnWithMsg(conn, fmt.Sprintf("Failed to handle buzz: %s", err), http.StatusInternalServerError)
				return
			}
		case RecvAns:
			var ansReq AnswerRequest
			if err := json.Unmarshal(msg, &ansReq); err != nil {
				log.Println("Failed to parse answer request:", err)
				closeConnWithMsg(conn, "Failed to parse answer request", http.StatusBadRequest)
			}
			playerId, err := getJWTSubject(ansReq.Token)
			if err != nil {
				log.Println("Failed to get playerId from token:", err)
				conn.Close()
				continue
			}
			err = game.handleAnswer(playerId, ansReq.Answer)
			if err != nil {
				log.Println("Failed to handle answer:", err)
				closeConnWithMsg(conn, fmt.Sprintf("Failed to handle answer: %s", err), http.StatusInternalServerError)
				return
			}
		case RecvWager:
			var wagerReq WagerRequest
			if err := json.Unmarshal(msg, &wagerReq); err != nil {
				log.Println("Failed to parse wager request:", err)
				closeConnWithMsg(conn, "Failed to parse wager request", http.StatusBadRequest)
				return
			}
			playerId, err := getJWTSubject(wagerReq.Token)
			if err != nil {
				log.Println("Failed to get playerId from token:", err)
				closeConnWithMsg(conn, "Failed to get playerId from token", http.StatusForbidden)
				return
			}
			err = game.handleWager(playerId, wagerReq.Wager)
			if err != nil {
				log.Println("Failed to handle wager:", err)
				closeConnWithMsg(conn, fmt.Sprintf("Failed to handle wager: %s", err), http.StatusInternalServerError)
				return
			}
		default:
			log.Println("Invalid game state")
			continue
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	if err := setJWTKeys(); err != nil {
		log.Fatalf("Failed to set JWT keys: %s", err)
	}

	r := gin.Default()
	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		log.Fatalf("Failed to set trusted proxies: %s", err)
	}
	r.Use(cors.Default())
	r.GET("/jeopardy/join", joinGame)
	r.GET("/jeopardy/play", playGame)

	addr := flag.String("addr", ":8080", "http service address")
	log.Fatal(r.Run(*addr))

}
