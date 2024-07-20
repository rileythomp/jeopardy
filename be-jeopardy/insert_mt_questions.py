import ast
import csv
import json
import os
import time

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
	create table if not exists mt_questions (
		id text,
		question text,
		answer text,
		alternatives text[],
		incorrect text[],
		source text,
		tags text[],
		other json
	);
	'''
)


def insert_file(file_path):
    with open(file_path, 'r') as csv_file:
        questions = csv.reader(csv_file)
        file_start = time.time()
        rows = 0
        for question in questions:
            rows += 1
            if rows == 1:
                continue
            question[3] = ast.literal_eval(question[3])
            question[4] = ast.literal_eval(question[4])
            question[6] = ast.literal_eval(question[6])
            question[7] = json.dumps(ast.literal_eval(question[7]))
            db.execute(
                '''
            	insert into mt_questions (id, question, answer, alternatives, incorrect, source, tags, other)
            	values (%s, %s, %s, %s, %s, %s, %s, %s)
            	''',
                question
            )
    print(
        f'Finished processing {file_path}, inserted {rows} rows in {time.time() - file_start:.2f} seconds')
    return rows


processed = 0
start = time.time()

with os.scandir('mt_questions') as files:
    for file in files:
        processed += insert_file(file.path)

print(f'Processed {processed} rows in {time.time() - start:.2f} seconds')

conn.commit()
db.close()
conn.close()
