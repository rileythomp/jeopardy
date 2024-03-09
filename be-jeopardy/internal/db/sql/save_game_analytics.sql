INSERT INTO jeopardy_analytics (
    game_id, created_at,
    first_round, first_round_ans, first_round_corr, first_round_score,
    second_round, second_round_ans, second_round_corr, second_round_score
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);