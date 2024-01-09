package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/auth"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/jeopardy"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/socket"
)

type (
	Route struct {
		Method  string
		Path    string
		Handler gin.HandlerFunc
	}

	JoinRequest struct {
		PlayerName string `json:"playerName"`
	}

	PlayRequest struct {
		Token string `json:"token,omitempty"`
	}
)

var (
	Routes = []Route{
		{
			Method:  "GET",
			Path:    "/jeopardy/health",
			Handler: CheckHealth,
		},
		{
			Method:  "GET",
			Path:    "/jeopardy/join",
			Handler: JoinGame,
		},
		{
			Method:  "GET",
			Path:    "/jeopardy/play",
			Handler: PlayGame,
		},
		{
			Method:  "GET",
			Path:    "/jeopardy/reset",
			Handler: TerminateGames,
		},
	}

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
			// return r.Header.Get("Origin") == "http://localhost:4200"
		},
	}
)

func JoinGame(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading connection to WebSocket: %s\n", err.Error())
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Error reading message from WebSocket: %s\n", err.Error())
		closeConnWithMsg(conn, "Error reading message from WebSocket", http.StatusInternalServerError)
		return
	}

	var joinReq JoinRequest
	if err := json.Unmarshal(msg, &joinReq); err != nil {
		log.Println("Error parsing join request:", err)
		closeConnWithMsg(conn, "Error parsing join request", http.StatusInternalServerError)
		return
	}

	game, playerId, err := jeopardy.JoinGame(joinReq.PlayerName)
	if err != nil {
		closeConnWithMsg(conn, "Error joining game", http.StatusInternalServerError)
		return
	}

	jwt, err := auth.GenerateJWT(playerId)
	if err != nil {
		log.Printf("Error generating JWT: %s\n", err.Error())
		closeConnWithMsg(conn, "Error generating JWT", http.StatusInternalServerError)
		return
	}

	resp := jeopardy.Response{
		Code:    200,
		Token:   jwt,
		Message: "Authorized to join game",
		Game:    game,
	}

	// TODO: close this connection
	err = conn.WriteJSON(resp)
	if err != nil {
		log.Printf("Error writing message to WebSocket: %s\n", err.Error())
		_ = conn.Close()
		return
	}
}

func PlayGame(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Error upgrading connection to WebSocket:", err)
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error reading message from WebSocket:", err)
		closeConnWithMsg(conn, "Error reading message WebSocket", http.StatusInternalServerError)
		return
	}

	var playReq PlayRequest
	if err := json.Unmarshal(msg, &playReq); err != nil {
		log.Println("Error unmarshalling play request:", err)
		closeConnWithMsg(conn, "Error parsing join request", http.StatusBadRequest)
		return
	}

	playerId, err := auth.GetJWTSubject(playReq.Token)
	if err != nil {
		log.Println("Error getting playerId from token:", err)
		closeConnWithMsg(conn, "Error getting playerId from token", http.StatusForbidden)
		return
	}

	safeConn := socket.NewSafeConn(conn)
	err = jeopardy.PlayGame(playerId, safeConn)
	if err != nil {
		log.Printf("Error during game: %s\n", err.Error())
		closeConnWithMsg(conn, "Error during game", http.StatusInternalServerError)
		return
	}
}

func TerminateGames(c *gin.Context) {
	log.Println("Closing all connections and terminating all games")
	jeopardy.TerminateGames()
	c.String(http.StatusOK, "Terminated all games")
}

func CheckHealth(c *gin.Context) {
	log.Println("Received health check")
	c.String(http.StatusOK, "OK")
}

func closeConnWithMsg(conn *websocket.Conn, msg string, code int) {
	_ = conn.WriteJSON(jeopardy.Response{Message: msg, Code: code})
	conn.Close()
}
