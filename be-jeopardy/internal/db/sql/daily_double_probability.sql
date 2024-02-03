with r1_prob as (
	select round, clue_value, concat(round(100*count(*)/sum(count(*)) over(), 2)::text, '%') as probability
	from jeopardy_clues
	where daily_double_value > 0
	group by round, clue_value
	having round = 1
	order by round asc, clue_value asc
),
r2_prob as (
	select round, clue_value, concat(round(100*count(*)/sum(count(*)) over(), 2)::text, '%') as probability
	from jeopardy_clues
	where daily_double_value > 0
	group by round, clue_value
	having round = 2
	order by round asc, clue_value asc
)
select * from r1_prob 
union 
select * from r2_prob
order by round asc, clue_value asc;