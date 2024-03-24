select games, wins, points, answers, correct, max_points, max_correct
from player_games
where email = $1;