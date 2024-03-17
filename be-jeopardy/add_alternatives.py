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
    add column if not exists alternatives text[];
    '''
)

db.execute(
    '''
    alter table jeopardy_clues 
    add column if not exists incorrect text[];
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
    set alternatives = array_append(
        alternatives,
        replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(replace(question, 'Á', 'A'), 'à', 'a'), 'á', 'a'), 'â', 'a'), 'ä', 'a'), 'ã', 'a'), 'å', 'a'), 'æ', 'a'), 'É', 'E'), 'è', 'e'), 'é', 'e'), 'ê', 'e'), 'ë', 'e'), 'ì', 'i'), 'í', 'i'), 'î', 'i'), 'ï', 'i'), 'Ö', 'O'), 'ò', 'o'), 'ó', 'o'), 'ø', 'o'), 'œ', 'o'), 'ô', 'o'), 'ö', 'o'), 'Ü', 'U'), 'ú', 'u'), 'û', 'u'), 'ü', 'u'), 'ñ', 'n'), 'ç', 'c'), 'ř', 'r')
    )
    where question like '%Á%' or
    question like '%à%' or
    question like '%á%' or
    question like '%â%' or
    question like '%ä%' or
    question like '%ã%' or
    question like '%å%' or
    question like '%æ%' or
    question like '%É%' or
    question like '%è%' or
    question like '%é%' or
    question like '%ê%' or
    question like '%ë%' or
    question like '%ì%' or
    question like '%í%' or
    question like '%î%' or
    question like '%ï%' or
    question like '%Ö%' or
    question like '%ò%' or
    question like '%ó%' or
    question like '%ø%' or
    question like '%œ%' or
    question like '%ô%' or
    question like '%ö%' or
    question like '%Ü%' or
    question like '%ú%' or
    question like '%û%' or
    question like '%ü%' or
    question like '%ñ%' or
    question like '%ç%' or
    question like '%ř%';
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
    set alternatives = array_append(alternatives, substring(question from 4))
    where question like 'an %';
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
