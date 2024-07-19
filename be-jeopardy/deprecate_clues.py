import psycopg2

conn = psycopg2.connect(
    dbname='postgres',
    user='postgres',
    password='postgres',
    host='localhost',
    port='5432'
)
db = conn.cursor()

db.execute(
    '''
    create table if not exists deprecated_clues (
        round int,
        clue_value int,
        daily_double_value int,
        category text,
        comments text,
        answer text,
        question text,
        air_date text,
        notes text,
        alternatives text[],
        incorrect text[]
    );
    '''
)

db.execute(
    '''
    insert into deprecated_clues (round, clue_value, daily_double_value, category, "comments", answer, question, air_date, notes, alternatives, incorrect) 
    select round, clue_value , daily_double_value, category, "comments", answer, question, air_date, notes, alternatives, incorrect 
    from jeopardy_clues
    where air_date like '2013%';
    '''
)

db.execute(
    '''
    delete from jeopardy_clues
    where air_date like '2013%';
    '''
)

print('Finished deprecating questions')

conn.commit()
db.close()
conn.close()
