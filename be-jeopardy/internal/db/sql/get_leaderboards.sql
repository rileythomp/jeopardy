	select email,
    wins, games, wins::float/games as win_rate,
    correct, answers, correct::float/answers as correct_rate,
    points, max_points, max_correct
	from player_games
	order by %s desc, correct_rate desc
	limit 10;