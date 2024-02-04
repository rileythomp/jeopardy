with r1_occurs as (
	select round, clue_value, round(10000*count(*)/sum(count(*)) over(), 0) as occurs
	from jeopardy_clues
	where daily_double_value > 0
	group by round, clue_value
	having round = 1
	order by round asc, clue_value asc
),
r2_occurs as (
	select round, clue_value, round(10000*count(*)/sum(count(*)) over(), 0) as occurs
	from jeopardy_clues
	where daily_double_value > 0
	group by round, clue_value
	having round = 2
	order by round asc, clue_value asc
),
r1_bounds as (
	select round, clue_value, occurs, sum(occurs) over(order by clue_value asc) as upper_bound
	from r1_occurs
),
r2_bounds as (
	select round, clue_value, occurs, sum(occurs) over(order by clue_value asc) as upper_bound
	from r2_occurs
)
select * from r1_bounds 
union 
select * from r2_bounds
order by round asc, clue_value asc;
