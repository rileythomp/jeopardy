package db

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/google/uuid"
)

type (
	AnalyticsRound struct {
		Categories []AnalyticsCategory
		Answers    *int
		Correct    *int
		Score      *float64
	}

	AnalyticsCategory struct {
		Title    string              `json:"title"`
		Question []AnalyticsQuestion `json:"question"`
	}

	AnalyticsQuestion struct {
		Answers []AnalyticsAnswer `json:"answers"`
	}

	AnalyticsAnswer struct {
		PlayerID    string `json:"playerId"`
		Answer      string `json:"answer"`
		Correct     bool   `json:"correct"`
		HasDisputed bool   `json:"hasDisputed"`
		Overturned  bool   `json:"overturned"`
		Bot         bool   `json:"bot"`
	}

	PlayerAnalytics struct {
		Games      int `json:"games"`
		Wins       int `json:"wins"`
		Points     int `json:"points"`
		Answers    int `json:"answers"`
		Correct    int `json:"correct"`
		MaxPoints  int `json:"maxPoints"`
		MaxCorrect int `json:"maxCorrect"`
	}

	LeaderboardUser struct {
		User
		PlayerAnalytics
		WinRate     float64 `json:"winRate"`
		CorrectRate float64 `json:"correctRate"`
	}
)

//go:embed sql/save_game_analytics.sql
var saveGameAnalytics string

func (db *JeopardyDB) SaveGameAnalytics(ctx context.Context, gameID uuid.UUID, createdAt int64, fr AnalyticsRound, sr AnalyticsRound) error {
	_, err := db.pool.Exec(
		ctx, saveGameAnalytics, gameID, createdAt,
		fr.Categories, fr.Answers, fr.Correct, fr.Score,
		sr.Categories, sr.Answers, sr.Correct, sr.Score,
	)
	return err
}

//go:embed sql/get_analytics.sql
var getAnalytics string

func (db *JeopardyDB) GetAnalytics(ctx context.Context) (any, error) {
	a := struct {
		GamesPlayed         int `json:"gamesPlayed"`
		FirstRoundAnsRate   int `json:"firstRoundAnsRate"`
		FirstRoundCorrRate  int `json:"firstRoundCorrRate"`
		FirstRoundScore     int `json:"firstRoundScore"`
		SecondRoundAnsRate  int `json:"secondRoundAnsRate"`
		SecondRoundCorrRate int `json:"secondRoundCorrRate"`
		SecondRoundScore    int `json:"secondRoundScore"`
	}{}
	err := db.pool.QueryRow(ctx, getAnalytics).Scan(
		&a.GamesPlayed,
		&a.FirstRoundAnsRate,
		&a.FirstRoundCorrRate,
		&a.FirstRoundScore,
		&a.SecondRoundAnsRate,
		&a.SecondRoundCorrRate,
		&a.SecondRoundScore,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

//go:embed sql/get_leaderboards.sql
var getLeaderboards string

func (db *JeopardyDB) GetLeaderboard(ctx context.Context, leaderboardType string) ([]*LeaderboardUser, error) {
	switch leaderboardType {
	case "win_rate", "wins", "games", "correct_rate", "correct", "answers", "points", "max_points", "max_correct":
	default:
		return nil, fmt.Errorf("invalid leaderboard type: %s", leaderboardType)
	}
	rows, err := db.pool.Query(ctx, fmt.Sprintf(getLeaderboards, leaderboardType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	leaderboard := []*LeaderboardUser{}
	for rows.Next() {
		user := &LeaderboardUser{}
		if err := rows.Scan(
			&user.Email,
			&user.Wins,
			&user.Games,
			&user.WinRate,
			&user.Correct,
			&user.Answers,
			&user.CorrectRate,
			&user.Points,
			&user.MaxPoints,
			&user.MaxCorrect,
		); err != nil {
			return nil, err
		}
		leaderboard = append(leaderboard, user)
	}
	return leaderboard, nil
}

//go:embed sql/get_player_analytics.sql
var getPlayerAnalytics string

func (db *JeopardyDB) GetPlayerAnalytics(ctx context.Context, email string) (PlayerAnalytics, error) {
	a := PlayerAnalytics{}
	err := db.pool.QueryRow(ctx, getPlayerAnalytics, email).Scan(
		&a.Games,
		&a.Wins,
		&a.Points,
		&a.Answers,
		&a.Correct,
		&a.MaxPoints,
		&a.MaxCorrect,
	)
	if err != nil {
		return PlayerAnalytics{}, err
	}
	return a, nil
}
