create table if not exists jeopardy_analytics (
    game_id uuid, 
    created_at bigint,
    first_round jsonb,
    first_round_ans int, 
    first_round_corr int,
    first_round_score double precision,
    second_round jsonb,
    second_round_ans int,
    second_round_corr int,
    second_round_score double precision
);