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
	}

	AnswerRequest struct {
		Request
		Answer string `json:"answer"`
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
		log.Println("Failed to unmarshal join request:", err)
		conn.Close()
		return
	}

	if game.readyToPlay() {
		log.Println("Too many players")
		_ = conn.WriteJSON(Response{Message: "Too many players", Code: http.StatusForbidden})
		conn.Close()
		return
	}

	playerId := game.addPlayer(joinReq.PlayerName)

	signedToken, err := generateToken(playerId)
	if err != nil {
		log.Println("Failed to generate token:", err)
		conn.Close()
		return
	}

	resp := Response{
		Token:   signedToken,
		Message: "Authorized to join game",
		Code:    200,
		Game:    game,
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
		log.Println("Failed to unmarshal play request:", err)
		conn.Close()
		return
	}

	playerId, err := getJWTSubject(playReq.Token)
	if err != nil {
		log.Println("Failed to get playerId from token:", err)
		conn.Close()
		return
	}

	player := game.getPlayerById(playerId)
	if player == nil {
		log.Println("Player not found")
		_ = conn.WriteJSON(Response{Message: "Player not found", Code: http.StatusForbidden})
		return
	}
	player.conn = conn

	readyToPlay := game.readyToPlay()

	var resp Response
	if readyToPlay {
		if err := game.setQuestions(); err != nil {
			log.Println("Failed to set questions:", err)
			conn.Close()
			return
		}
		game.setState(RecvPick, game.Players[0].Id)
		resp = Response{
			Code:    200,
			Message: "We are ready to play",
			Game:    game,
		}
	} else {
		playersReady := game.numPlayersReady()
		resp = Response{
			Code:    200,
			Message: fmt.Sprintf("There are %d players ready, waiting for %d more", playersReady, 3-playersReady),
			Game:    game,
		}
	}

	if err := game.messageAllPlayers(resp); err != nil {
		log.Println("Error sending message to players:", err)
		conn.Close()
		return
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read message from WebSocket:", err)
			conn.Close()
			return
		}
		if game.State == RecvPick {
			var pickReq PickRequest
			if err := json.Unmarshal(msg, &pickReq); err != nil {
				log.Println("Failed to unmarshal pick request:", err)
				conn.Close()
				continue
			}
			playerId, err := getJWTSubject(pickReq.Token)
			if err != nil {
				log.Println("Failed to get playerId from token:", err)
				conn.Close()
				continue
			}
			player := game.getPlayerById(playerId)
			if player == nil {
				log.Println("Player not found")
				_ = conn.WriteJSON(Response{Message: "Player not found", Code: http.StatusForbidden})
				continue
			}
			if !player.CanPick {
				log.Println("Player cannot pick")
				_ = conn.WriteJSON(Response{Message: "Player cannot pick", Code: http.StatusForbidden})
				continue
			}
			curQuestion := game.FirstRound[pickReq.TopicIdx].Questions[pickReq.ValIdx]
			if !curQuestion.CanChoose {
				log.Println("Question already chosen")
				_ = conn.WriteJSON(Response{Message: "Question already chosen", Code: http.StatusForbidden})
				continue
			}
			game.LastPicker = player.Id
			game.CurQuestion = curQuestion
			game.setState(RecvBuzz, "")
			resp := Response{
				Code:    200,
				Message: "New Question",
				Game:    game,
			}
			if err := game.messageAllPlayers(resp); err != nil {
				log.Println("Error sending question to players:", err)
				conn.Close()
				continue
			}
		} else if game.State == RecvBuzz {
			var buzzReq BuzzRequest
			if err := json.Unmarshal(msg, &buzzReq); err != nil {
				log.Println("Failed to unmarshal buzz request:", err)
				conn.Close()
				continue
			}
			playerId, err := getJWTSubject(buzzReq.Token)
			if err != nil {
				log.Println("Failed to get playerId from token:", err)
				conn.Close()
				continue
			}
			player := game.getPlayerById(playerId)
			if player == nil {
				log.Println("Player not found")
				_ = conn.WriteJSON(Response{Message: "Player not found", Code: http.StatusForbidden})
				continue
			}
			if !player.CanBuzz {
				log.Println("Player cannot buzz")
				_ = conn.WriteJSON(Response{Message: "Player cannot buzz", Code: http.StatusForbidden})
				continue
			}
			player.CanBuzz = false
			game.setState(RecvAns, player.Id)
			resp := Response{
				Code:    200,
				Message: "Player buzzed",
				Game:    game,
			}
			if err := game.messageAllPlayers(resp); err != nil {
				log.Println("Error sending buzz to players:", err)
				conn.Close()
				continue
			}
		} else if game.State == RecvAns {
			var ansReq AnswerRequest
			if err := json.Unmarshal(msg, &ansReq); err != nil {
				log.Println("Failed to unmarshal buzz request:", err)
				conn.Close()
				continue
			}
			playerId, err := getJWTSubject(ansReq.Token)
			if err != nil {
				log.Println("Failed to get playerId from token:", err)
				conn.Close()
				continue
			}
			player := game.getPlayerById(playerId)
			if player == nil {
				log.Println("Player not found")
				_ = conn.WriteJSON(Response{Message: "Player not found", Code: http.StatusForbidden})
				continue
			}
			if !player.CanAnswer {
				log.Println("Player cannot answer")
				_ = conn.WriteJSON(Response{Message: "Player cannot answer", Code: http.StatusForbidden})
				continue
			}
			isCorrect := game.CurQuestion.checkAnswer(ansReq.Answer)
			if isCorrect {
				player.Score += game.CurQuestion.Value
				game.disableQuestion(game.CurQuestion)
				game.CurQuestion = Question{}
				game.setState(RecvPick, player.Id)
				resp := Response{
					Code:    200,
					Message: "Player answered correctly",
					Game:    game,
				}
				if err := game.messageAllPlayers(resp); err != nil {
					log.Println("Error sending correct answer to players:", err)
					conn.Close()
					continue
				}
				game.GuessedWrong = []string{}
			} else {
				player.Score -= game.CurQuestion.Value
				game.GuessedWrong = append(game.GuessedWrong, player.Id)
				allGuessed := len(game.GuessedWrong) == len(game.Players)
				if allGuessed {
					game.disableQuestion(game.CurQuestion)
					game.setState(RecvPick, game.LastPicker)
				} else {
					game.setState(RecvBuzz, "")
				}
				resp := Response{
					Code:    200,
					Message: "Player answered incorrectly",
					Game:    game,
				}
				if err := game.messageAllPlayers(resp); err != nil {
					log.Println("Error sending incorrect answer to players:", err)
					conn.Close()
					continue
				}
				if allGuessed {
					game.GuessedWrong = []string{}
				}
			}
		} else {
			continue
		}

	}
}

func resetGame(c *gin.Context) {
	game = NewGame()
	c.JSON(http.StatusOK, gin.H{"message": "Game reset"})
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
	r.GET("/jeopardy/reset", resetGame)
	log.Fatal(r.Run(*addr))

}
