update jeopardy_clues 
set incorrect = array_append(incorrect, $1)
where answer = $2;