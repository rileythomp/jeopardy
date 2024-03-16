select category, round, air_date
from jeopardy_clues 
where lower(category) like concat($3::text, $1::text, '%')
group by category, round, air_date
having (round = 1 or round = $2) and count(*) = 5
order by category asc;