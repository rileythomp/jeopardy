insert into player_games (email, games, wins, points, answers, correct, max_points, max_correct)
values ($1, 1, $2, $3, $4, $5, $3, $5)
on conflict (email)
do update set 
games = player_games.games + 1,
wins = player_games.wins + $2,
points = player_games.points + $3,
answers = player_games.answers + $4,
correct = player_games.correct + $5,
max_points = greatest(player_games.max_points, $3),
max_correct = greatest(player_games.max_correct, $5)
where player_games.email = $1;
