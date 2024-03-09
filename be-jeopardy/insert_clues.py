# NOTE: Run remove_slashes.sql manually after this script.
# NOTE: Run add_alternatives.py after remove_slashes.sql

import csv
import psycopg2
import os
import time

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
    create table if not exists jeopardy_clues (
        round int,
        clue_value int,  
        daily_double_value int,
        category text, 
        comments text,
        answer text, 
        question text, 
        air_date text, 
        notes text
    );
    '''
)

db.execute(
    '''
    create table if not exists jeopardy_analytics (
        game_id uuid, 
        created_at bigint,
        first_round jsonb,
        first_round_ans int, 
        first_round_corr int,
        second_round jsonb,
        second_round_ans int,
        second_round_corr int
    );
    '''
)

inserts = 0
start = time.time()
with os.scandir('clues') as files:
    for file in files:
        print(f'Starting to process {file.path}')
        rows = 0
        file_start = time.time()
        with open(file.path, 'r') as tsv:
            clues = csv.reader(tsv, delimiter='\t')
            for clue in clues:
                rows += 1
                if rows == 1:
                    continue
                db.execute(
                    'insert into jeopardy_clues (round, clue_value, daily_double_value, category, comments, answer, question, air_date, notes) values (%s, %s, %s, %s, %s, %s, %s, %s, %s)',
                    clue
                )
                inserts += 1
                if rows % 1000 == 0:
                    print(f'Inserted {rows} rows')
        file_end = time.time()
        print(
            f'Finished processing {file.path}, inserted {rows} rows in {file_end - file_start:.2f} seconds\n')

db.execute(
    '''
    update jeopardy_clues 
    set category = replace(category, '\\"', '"'),
    answer = replace(answer, '\\"', '"'),
    question = replace(question, '\\"', '"');
    '''
)

db.execute(
    '''
    update jeopardy_clues 
    set comments = ''
    where comments != '' and comments not like '(%: %)';
    '''
)

db.execute(
    '''
    update jeopardy_clues 
    set comments = rtrim(substring(comments from position(':' in comments) + 2), ')') 
    where comments != '' and comments like '(%: %)';
    '''
)

end = time.time()
print(
    f'Finished processing all files, inserted {inserts} rows in {end - start:.2f} seconds')

conn.commit()
db.close()
conn.close()

# NOTE: Run remove_slashes.sql manually after this script.
# NOTE: Run add_alternative.py after remove_slashes.sql
