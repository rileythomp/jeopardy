select
count(*),
round(avg(100.0*first_round_ans/30), 0) as first_round_answer_rate,
round(avg(100.0*first_round_corr/first_round_ans), 0) as first_round_correctness_rate,
round(avg(first_round_score)::numeric, 0) as first_round_score,
round(avg(100.0*second_round_ans/30), 0) as second_round_answer_rate,
round(avg(100.0*second_round_corr/second_round_ans), 0) as second_round_correctness_rate,
round(avg(second_round_score)::numeric, 0) as second_round_score
from jeopardy_analytics;