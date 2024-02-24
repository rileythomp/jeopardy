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
    alter table jeopardy_clues 
    add column alternatives text[];
    '''
)

db.execute(
    '''
    update jeopardy_clues 
    set alternatives = ARRAY[question];
    '''
)

db.execute(
    '''
    update jeopardy_clues 
    set alternatives = array_append(alternatives, substring(question from 5))
    where question like 'the %';
    '''
)

db.execute(
    '''
    update jeopardy_clues 
    set alternatives = array_append(alternatives, substring(question from 3))
    where question like 'a %';
    '''
)

db.execute(
    '''
    update jeopardy_clues 
    set alternatives = array_append(alternatives,  trim(regexp_replace(regexp_replace(question, '\(.*\)', '', 'g'), '  ', ' ', 'g'))) 
    where question like '%(%)%';
    '''
)

db.execute(
    '''
    update jeopardy_clues 
    set alternatives = array_append(alternatives, substring(trim(regexp_replace(regexp_replace(question, '\(.*\)', '', 'g'), '  ', ' ', 'g')) from 5))
    where question like 'the %(%)%';
    '''
)


db.execute(
    '''
    update jeopardy_clues 
    set alternatives = array_append(alternatives, substring(trim(regexp_replace(regexp_replace(question, '\(.*\)', '', 'g'), '  ', ' ', 'g')) from 3))
    where question like 'a %(%)%';
    '''
)

db.execute(
    '''
    update jeopardy_clues 
    set alternatives = array_append(alternatives, trim(regexp_replace(regexp_replace(question, '\".*\"', '', 'g'), '  ', ' ', 'g')))
    where question like '%"The Rock"%'
    or question like '%"Bones"%'
    or question like '%"Fatty"%'
    or question like '%"Easy"%'
    or question like '%"Capability"%'
    or question like '%"The Gipper"%';
    '''
)

conn.commit()
db.close()
conn.close()
