import csv
import re
from datetime import datetime

import requests
from bs4 import BeautifulSoup


class JeopardyQuestion:
    def __init__(self, rnd, clue_value, daily_double_value,
                 category, comments, answer, question, air_date, notes):
        self.round = rnd
        self.clue_value = clue_value
        self.daily_double_value = daily_double_value
        self.category = category
        self.comments = comments
        self.answer = answer
        self.question = question
        self.air_date = air_date
        self.notes = notes


url = 'https://j-archive.com/showseason.php?season=40'

response = requests.get(url)

if response.status_code != 200:
    print(response)
    print('Failed to fetch the page.')
    exit()

soup = BeautifulSoup(response.text, 'html.parser')

tables = soup.find_all('table')
assert (len(tables) == 1)

rows = tables[0].find_all('tr')
print('number of games', len(rows))

questions = []

games = 1
for row in rows:
    print(f'{games}/{len(rows)}')
    games += 1
    tds = row.find_all('td')
    assert (len(tds) == 3)
    path = tds[0].find('a').get('href')
    print(path)
    game_url = f'https://j-archive.com/{path}'
    resp = requests.get(game_url)
    if resp.status_code != 200:
        print(resp)
        print(f'Failed to fetch the game page: {game_url}')
        exit()
    game_soup = BeautifulSoup(resp.text, 'html.parser')
    round_tables = game_soup.find_all('table', class_='round')
    assert (len(round_tables) == 2)
    final_tables = game_soup.find_all('table', class_='final_round')
    assert (len(final_tables) < 3)
    air_date = datetime.strptime(game_soup.find(
        id='game_title').text.split(' - ')[1], "%A, %B %d, %Y").strftime("%Y-%m-%d")

    rnd_num = 1
    for rnd in round_tables:
        categories = [c.text for c in rnd.find_all(
            'td', class_='category_name')]
        category_comments = [c.text for c in rnd.find_all(
            'td', class_='category_comments')]
        assert (len(categories) == len(category_comments) == 6)
        clues = rnd.find_all('td', class_='clue')
        col = 0
        for clue in clues:
            val = clue.find('td', class_='clue_value')
            daily_double = False
            if val is None:
                val = clue.find('td', class_='clue_value_daily_double')
                if val is None:
                    continue
                val = int("".join(re.findall(r'\d+', val.text)))
                daily_double = True
            else:
                val = int("".join(re.findall(r'\d+', val.text)))
            answer, question = clue.find_all('td', class_='clue_text')
            questions.append({
                'round': rnd_num,
                'clue_value': val,
                'daily_double_value': val if daily_double else 0,
                'category': categories[col],
                'comments': category_comments[col],
                'answer': answer.text,
                'question': question.find('em', class_='correct_response').text,
                'air_date': air_date,
                'notes': ''
            })
            col = (col+1) % 6
        rnd_num += 1

    for ft in final_tables:
        answer, question = ft.find_all('td', class_='clue_text')
        questions.append({
            'round': 3,
            'clue_value': 0,
            'daily_double_value': 0,
            'category': ft.find('td', class_='category_name').text,
            'comments': ft.find('td', class_='category_comments').text,
            'answer': answer.text,
            'question': question.find('em', class_='correct_response').text,
            'air_date': air_date,
            'notes': ''
        })

    print()

with open('season40.tsv', 'w') as file:
    csv_writer = csv.writer(file, delimiter='\t')
    headers = ['round', 'clue_value', 'daily_double_value', 'category',
               'comments', 'answer', 'question', 'air_date', 'notes']
    csv_writer.writerow(headers)
    csv_writer.writerows([q.values() for q in questions])
