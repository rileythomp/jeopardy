package jeopardy

type (
	Response struct {
		Code      int     `json:"code"`
		Token     string  `json:"token,omitempty"`
		Message   string  `json:"message"`
		Game      *Game   `json:"game,omitempty"`
		CurPlayer *Player `json:"curPlayer,omitempty"`
	}

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

	ConfirmAnsRequest struct {
		Request
		Confirm bool `json:"confirm"`
	}

	WagerRequest struct {
		Request
		Wager int `json:"wager"`
	}
)
