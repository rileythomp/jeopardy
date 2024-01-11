package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/auth"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/jeopardy"
	"github.com/rileythomp/jeopardy/be-jeopardy/internal/log"
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
		GameName   string `json:"gameName"`
		Private    bool   `json:"private"`
	}

	PlayRequest struct {
		Token string `json:"token,omitempty"`
	}
)

var (
	Routes = []Route{
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/health",
			Handler: CheckHealth,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/join",
			Handler: JoinGame,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/play",
			Handler: PlayGame,
		},
		{
			Method:  http.MethodGet,
			Path:    "/jeopardy/games",
			Handler: GetGames,
		},
	}

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
			// TODO: SET THIS PROPERLY
			// return r.Header.Get("Origin") == "http://localhost:4200"
		},
	}
)

func JoinGame(c *gin.Context) {
	log.Infof("Received join request")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("Error upgrading connection to WebSocket: %s\n", err.Error())
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Errorf("Error reading message from WebSocket: %s\n", err.Error())
		closeConnWithMsg(conn, "Error reading message from WebSocket", http.StatusInternalServerError)
		return
	}

	var req JoinRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		log.Errorf("Error parsing join request: %s\n", err.Error())
		closeConnWithMsg(conn, "Error parsing join request", http.StatusInternalServerError)
		return
	}

	game, playerId, err := jeopardy.JoinGame(req.PlayerName, req.GameName, req.Private)
	if err != nil {
		closeConnWithMsg(conn, "Error joining game", http.StatusInternalServerError)
		return
	}

	jwt, err := auth.GenerateJWT(playerId)
	if err != nil {
		log.Errorf("Error generating JWT: %s\n", err.Error())
		closeConnWithMsg(conn, "Error generating JWT", http.StatusInternalServerError)
		return
	}

	if err = conn.WriteJSON(jeopardy.Response{
		Code:    200,
		Token:   jwt,
		Message: "Authorized to join game",
		Game:    game,
	}); err != nil {
		log.Errorf("Error writing message to WebSocket: %s\n", err.Error())
		return
	}
	if err = conn.Close(); err != nil {
		log.Errorf("Error closing WebSocket: %s\n", err.Error())
		return
	}
}

func PlayGame(c *gin.Context) {
	log.Infof("Received play request")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("Error upgrading connection to WebSocket: %s\n", err.Error())
		c.Writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, msg, err := conn.ReadMessage()
	if err != nil {
		log.Errorf("Error reading message from WebSocket: %s\n", err.Error())
		closeConnWithMsg(conn, "Error reading message WebSocket", http.StatusInternalServerError)
		return
	}

	var req PlayRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		log.Errorf("Error unmarshalling play request: %s\n", err.Error())
		closeConnWithMsg(conn, "Error parsing join request", http.StatusBadRequest)
		return
	}

	playerId, err := auth.GetJWTSubject(req.Token)
	if err != nil {
		log.Errorf("Error getting playerId from token: %s\n", err.Error())
		closeConnWithMsg(conn, "Error getting playerId from token", http.StatusForbidden)
		return
	}

	safeConn := socket.NewSafeConn(conn)
	err = jeopardy.PlayGame(playerId, safeConn)
	if err != nil {
		log.Errorf("Error during game: %s\n", err.Error())
		closeConnWithMsg(conn, "Error during game", http.StatusInternalServerError)
		return
	}
}

func GetGames(c *gin.Context) {
	log.Infof("Received request to get games")
	games := jeopardy.GetPrivateGames()
	c.JSON(http.StatusOK, games)
}

func CheckHealth(c *gin.Context) {
	log.Infof("Received health check")
	c.String(http.StatusOK, "OK")
}

func closeConnWithMsg(conn *websocket.Conn, msg string, code int) {
	_ = conn.WriteJSON(jeopardy.Response{Message: msg, Code: code})
	conn.Close()
}
