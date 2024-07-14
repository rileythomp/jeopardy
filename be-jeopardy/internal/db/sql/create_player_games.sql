create table if not exists player_games (
    email text primary key, 
    games int,
    wins int,
    points int,
    answers int,
    correct int,
    max_points int,
    max_correct int
);