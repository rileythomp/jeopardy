package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/jeopardy"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
			// return r.Header.Get("Origin") == "http://localhost:4200"
		},
	}

	playerGames = map[string]*jeopardy.Game{}
	games       = []*jeopardy.Game{}
)

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

	r.GET("/jeopardy/health", checkHealth)
	r.GET("/jeopardy/join", joinGame)
	r.GET("/jeopardy/play", playGame)
	r.GET("/jeopardy/reset", terminateGames)

	port := os.Getenv("PORT")
	addr := flag.String("addr", ":"+port, "http service address")
	log.Fatal(r.Run(*addr))

}

func closeConnWithMsg(conn *websocket.Conn, msg string, code int) {
	_ = conn.WriteJSON(jeopardy.Response{Message: msg, Code: code})
	conn.Close()
}

func terminateGames(c *gin.Context) {
	log.Println("Closing all connections and terminating all games")

	for i, game := range games {
		log.Printf("Terminating game %d\n", i)
		game.TerminateGame()
	}

	playerGames = map[string]*jeopardy.Game{}
	games = []*jeopardy.Game{}

	c.String(http.StatusOK, "Terminated all games")
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

	var game = jeopardy.NewGame()
	for _, g := range games {
		if len(g.Players) < 3 {
			game = g
		}
	}
	playerId, err := game.AddPlayer(joinReq.PlayerName)
	if err != nil {
		log.Println("Failed to add player:", err)
		closeConnWithMsg(conn, fmt.Sprintf("Failed to add player: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	playerGames[playerId] = game
	if len(game.Players) == 1 {
		games = append(games, game)
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

	game := playerGames[playerId]
	err = game.SetPlayerConnection(playerId, conn)
	if err != nil {
		log.Println("Failed to set player connection:", err)
		closeConnWithMsg(conn, "Failed to set player connection", http.StatusInternalServerError)
		return
	}

	// TODO: USE A CHANNEL TO WAIT ON A MESSAGE OR TO END THE GAME
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

func checkHealth(c *gin.Context) {
	log.Println("Received health check")
	c.String(http.StatusOK, "OK")
}
