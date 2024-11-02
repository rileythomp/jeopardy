import csv
import json
import re
import sys
import time
import uuid
from typing import List, Set

import requests
from bs4 import BeautifulSoup


def output_url_json(url: str, output: str = None):
    response = requests.get(url)
    if response.status_code != 200:
        print('Failed to fetch the page.')
        return

    content = response.text
    match = re.search(r'var _page = (.*);', content)
    if not match:
        print('Could not find the JSON object.')
        return

    json_obj = json.loads(match.group(1))
    print(json.dumps(json_obj, indent=4), file=output)


def get_url_json(url: str):
    response = requests.get(url)
    if response.status_code != 200:
        print('Failed to fetch the page.')
        return

    content = response.text
    match = re.search(r'var _page = (.*);', content)
    if not match:
        print('Could not find the JSON object.')
        return

    return json.loads(match.group(1))


def clean_html_string(html_string):
    # put a space between html tags, remove all html tags, remove extra whitespace
    return re.sub(r'\s+', ' ', re.sub(r'<[^>]+>', '', re.sub(r'>\s*<', '> <', html_string)).strip())


class JetpunkQuestion:
    def __init__(self, question, stats, rating, url: str, tags: Set[str]):
        self.id = uuid.uuid4()
        self.question = clean_html_string(question['cols'][0])
        self.answer = re.sub(
            r'{|}', '', clean_html_string(question['cols'][1]))
        self.alternatives = [typein['val']
                             for typein in question.get('typeins', [])]
        self.source = 'www.jetpunk.com'
        self.tags = list(tags)
        self.other = {
            'pct': stats['pct'],
            'url': url,
            'rating': rating,
        }
        self.incorrect = []
        assert (question['cols'][1] == question['display'])
        assert (question['id'] == stats['id'])
        assert (question['cols'][1] == stats['display'])
        stats_typeins = stats['typeins'] if stats['typeins'] is not None else [
        ]
        assert (self.alternatives == [typein['val']
                for typein in stats_typeins])

    def output(self, output=sys.stdout):
        print(self.id, self.question, self.answer, self.alternatives,
              self.incorrect, self.source, self.tags, self.other, file=output)

    def csv_output(self, writer):
        writer.writerow([self.id, self.question, self.answer, self.alternatives,
                        self.incorrect, self.source, self.tags, self.other])


def jetpunk_quizzes_to_csv(quiz_urls: List[str], csv_name: str):
    with open(csv_name, 'w') as file:
        csv_writer = csv.writer(file)
        headers = ['id', 'question', 'answer', 'alternatives',
                   'incorrect', 'source', 'tags', 'other']
        csv_writer.writerow(headers)
        i = 1
        start = time.time()
        for url in quiz_urls:
            print(i, time.strftime('%H:%M:%S', time.gmtime(time.time() - start)))
            i += 1
            print('Starting', url)
            quiz_json = get_url_json(url)
            stats_json = get_url_json(f'{url}/stats')

            quiz = quiz_json['data']['quiz']
            stats = stats_json['data']['svgStats']

            assert (quiz['id'] == stats_json['data']['quiz']['id'])
            if len(quiz['answers']) != len(stats):
                print('Different number of questions and answers for', url)

            if quiz['whatkind'] == 'mc':
                print('Skipping multiple choice quiz', url)
                print()
                continue
            if quiz['whatkind'] == 'ts':
                print('Skipping tile select in quiz', url)
                print()
                continue

            tags: Set[str] = set(["general-knowledge"])
            quiz_url = quiz['url']
            match = re.match(
                r"(.+?)(?:-general-knowledge-quiz|-quiz|-knowledge|-general-knowledge)(?:-\d+)?$", quiz_url)
            if match:
                tags.add(match.group(1))

            broke = False
            questions = quiz['answers']
            for q in questions:
                try:
                    question = JetpunkQuestion(
                        q, stats[q['id']], quiz['rating'], url, tags)
                    question.csv_output(csv_writer)
                except (AssertionError, IndexError) as e:
                    print(f'Failed to parse question: {e}')
                    broke = True
                    break
            if broke:
                print('Skipping due to parsing error', url)
                print()
                continue

            print(f'Wrote {url} to {csv_name}')
            print()


if __name__ == '__main__':
    # quiz_urls = [f'https://www.jetpunk.com/quizzes/general-knowledge-quiz-{i}' for i in range(1, 230)]
    # jetpunk_quizzes_to_csv(quiz_urls, 'general_knowledge.csv')
    # jetpunk_quizzes_to_csv(['https://www.jetpunk.com/quizzes/general-knowledge-think-fast'], 'tmp.csv')

    quizzes = []
    for i in range(1, 4):
        response = requests.get(
            f'https://www.jetpunk.com/search?term=General%20Knowledge&language=english&page={i}')

        if response.status_code != 200:
            print(
                f"Failed to retrieve the webpage: Status code {response.status_code}")
            exit()

        soup = BeautifulSoup(response.text, 'html.parser')

        super_tables = soup.find_all('table', class_='super-table')
        assert (len(super_tables) == 1)

        super_table = super_tables[0]
        for tr in super_table.find_all('tr'):
            if tr.find('th'):
                continue
            tds = tr.find_all('td')
            assert (len(tds) == 2)
            td = tds[1]
            if td.find('i'):
                link = td.find('a')
                path = link.get('href')
                title = link.get_text()
                quizzes.append({'path': path, 'title': title})

    quizzes.sort(key=lambda x: x['title'])

    quiz_urls = [f"https://www.jetpunk.com{quiz['path']}" for quiz in quizzes]
    jetpunk_quizzes_to_csv(quiz_urls, 'jetpunk.csv')
