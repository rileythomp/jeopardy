update jeopardy_clues 
set alternatives = array_append(alternatives, $1)
where question = $2;