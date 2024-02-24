-- NOTE: Run add_alternatives.py after this script

UPDATE jeopardy_clues 
SET category = REGEXP_REPLACE(category, '\\''+', '''', 'g'),
answer = REGEXP_REPLACE(answer, '\\''+', '''', 'g'),
question = REGEXP_REPLACE(question,  '\\''+', '''', 'g');

-- NOTE: Run add_alternatives.py after this script
