select round, clue_value, category, comments, answer, question, alternatives 
from jeopardy_clues
where category = $1 and air_date = $2 and round = $3;
