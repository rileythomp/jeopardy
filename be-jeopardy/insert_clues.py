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
                del clue[2]
                db.execute(
                    'insert into jeopardy_clues (round, clue_value, category, comments, answer, question, air_date, notes) values (%s, %s, %s, %s, %s, %s, %s, %s)',
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

end = time.time()
print(
    f'Finished processing all files, inserted {inserts} rows in {end - start:.2f} seconds')

conn.commit()
db.close()
conn.close()
