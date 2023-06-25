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
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/jeopardy"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get("Origin") == "http://localhost:4200"
		},
	}

	game = jeopardy.NewGame()
)

func closeConnWithMsg(conn *websocket.Conn, msg string, code int) {
	_ = conn.WriteJSON(jeopardy.Response{Message: msg, Code: code})
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
	var joinReq jeopardy.JoinRequest
	if err := json.Unmarshal(msg, &joinReq); err != nil {
		log.Println("Failed to parse join request:", err)
		closeConnWithMsg(conn, fmt.Sprintf("Failed to parse join request: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	playerId, err := game.AddPlayer(joinReq.PlayerName)
	if err != nil {
		log.Println("Failed to add player:", err)
		closeConnWithMsg(conn, fmt.Sprintf("Failed to add player: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	jwt, err := generateJWT(playerId)
	if err != nil {
		log.Println("Failed to generate token:", err)
		closeConnWithMsg(conn, "Failed to generate token", http.StatusInternalServerError)
		return
	}
	resp := jeopardy.Response{
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
	var playReq jeopardy.PlayRequest
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

	err = game.SetPlayerConnection(playerId, conn)
	if err != nil {
		log.Println("Failed to set player connection:", err)
		closeConnWithMsg(conn, "Failed to set player connection", http.StatusInternalServerError)
		return
	}

	resp := jeopardy.Response{
		Code:    200,
		Message: "Waiting for more players",
		Game:    game,
	}
	if game.ReadyToPlay() {
		if err := game.StartGame(); err != nil {
			log.Println("Error starting game", err)
			closeConnWithMsg(conn, fmt.Sprintf("error starting game: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		resp = jeopardy.Response{
			Code:    200,
			Message: "We are ready to play",
			Game:    game,
		}
	}

	if err := game.MessageAllPlayers(resp); err != nil {
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
		var req jeopardy.Request
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Println("Failed to parse request:", err)
			closeConnWithMsg(conn, "Failed to parse request", http.StatusBadRequest)
			return
		}
		playerId, err := getJWTSubject(req.Token)
		if err != nil {
			log.Println("Failed to get playerId from token:", err)
			closeConnWithMsg(conn, "Failed to get playerId from token", http.StatusForbidden)
			return
		}
		err = game.HandleRequest(playerId, msg)
		if err != nil {
			log.Printf("Failed to handle request: %s", err.Error())
			closeConnWithMsg(conn, err.Error(), http.StatusInternalServerError)
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
