package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type (
	JoinRequest struct {
		PlayerName string `json:"playerName"`
	}

	PlayRequest struct {
		Token string `json:"token,omitempty"`
	}

	Response struct {
		Code    int    `json:"code"`
		Token   string `json:"token,omitempty"`
		Message string `json:"message"`
		Game    *Game  `json:"game,omitempty"`
	}
)

var (
	addr = flag.String("addr", ":8080", "http service address")

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == "http://localhost:4200"
		},
	}

	game = NewGame()
)

func joinGame(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade connection to WebSocket:", err)
		conn.Close()
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Failed to read message from WebSocket:", err)
		conn.Close()
		return
	}

	var joinReq JoinRequest
	if err := json.Unmarshal(msg, &joinReq); err != nil {
		log.Println("Failed to unmarshal JSON:", err)
		conn.Close()
		return
	}

	if len(game.Players) > 2 {
		log.Println("Too many players")
		_ = conn.WriteJSON(Response{Message: "Too many players", Code: http.StatusForbidden})
		conn.Close()
		return
	}

	player := &Player{
		Id:   uuid.New().String(),
		Name: joinReq.PlayerName,
	}
	game.Players[player.Id] = player

	signedToken, err := generateToken(player.Id)
	if err != nil {
		log.Println("Failed to generate token:", err)
		conn.Close()
		return
	}

	resp := Response{
		Token:   signedToken,
		Message: "Authorized to join game",
		Code:    200,
	}

	err = conn.WriteJSON(resp)
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
		conn.Close()
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Failed to read message from WebSocket:", err)
		conn.Close()
		return
	}

	var playReq PlayRequest
	if err := json.Unmarshal(msg, &playReq); err != nil {
		log.Println("Failed to unmarshal JSON:", err)
		conn.Close()
		return
	}

	playerId, err := getJWTSubject(playReq.Token)
	if err != nil {
		log.Println("Failed to get playerId from token:", err)
		conn.Close()
		return
	}

	player, ok := game.Players[playerId]
	if !ok {
		log.Println("Player not found")
		_ = conn.WriteJSON(Response{Message: "Player not found", Code: http.StatusForbidden})
		return
	}
	player.conn = conn

	resp := Response{
		Code:    200,
		Message: "We are ready to play",
		Game:    game,
	}

	playersReady := game.numPlayersReady()
	if playersReady < 3 {
		resp = Response{
			Code:    200,
			Message: fmt.Sprintf("There are %d players ready, waiting for %d more", playersReady, 3-playersReady),
			Game:    game,
		}
	}

	if err := game.messagePlayers(resp); err != nil {
		log.Println("Error sending message to players:", err)
		conn.Close()
		return
	}

	for {
		// read client message
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read message from WebSocket:", err)
			conn.Close()
			return
		}
		type Tmp struct {
			Field string `json:"hello"`
			Echo  string `json:"echo"`
		}
		var t Tmp
		if err := json.Unmarshal(msg, &t); err != nil {
			log.Println("Failed to unmarshal JSON:", err)
			conn.Close()
			return
		}
		t.Echo = t.Field + "addition"

		err = conn.WriteJSON(t)
		if err != nil {
			log.Println("Failed to write message to WebSocket:", err)
			conn.Close()
			return
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
	log.Fatal(r.Run(*addr))

}
