select
count(*) over() as games_played,
round(100.0*sum(first_round_ans) over()/30, 0) as first_round_answer_rate,
round(100.0*sum(first_round_corr) over()/sum(first_round_ans) over(), 0) as first_round_correctness_rate,
round(100.0*sum(second_round_ans) over()/30, 0) as second_round_answer_rate,
round(100.0*sum(second_round_corr) over()/sum(second_round_ans) over(), 0) as second_round_correctness_rate
from jeopardy_analytics;