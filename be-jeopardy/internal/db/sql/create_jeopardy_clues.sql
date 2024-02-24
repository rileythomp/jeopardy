create table if not exists jeopardy_clues (
	round int,
	clue_value int,  
	daily_double_value int,
	category text, 
	comments text,
	answer text, 
	question text, 
	air_date text, 
	notes text,
	alternatives text[]
);