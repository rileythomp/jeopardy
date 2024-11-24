import csv
import json
import sys
import time
import uuid

import requests


class TheTriviaAPIQuestion:
    def __init__(self, question):
        self.id = uuid.uuid4()
        self.question = question['question']['text']
        self.answer = question['correctAnswer']
        self.alternatives = []
        self.source = 'the-trivia-api.com/v2'
        tags = set(question['tags'])
        tags.add(question['category'])
        self.tags = list(tags)
        self.other = {
            'id': question['id'],
            'difficulty': question['difficulty'],
        }
        self.incorrect = question['incorrectAnswers']

    def output(self, output=sys.stdout):
        print(self.id, self.question, self.answer, self.alternatives,
              self.incorrect, self.source, self.tags, self.other, file=output)

    def csv_output(self, writer):
        writer.writerow([self.id, self.question, self.answer, self.alternatives,
                        self.incorrect, self.source, self.tags, self.other])


def get_url_json(url: str):
    response = requests.get(url)
    if response.status_code != 200:
        print('Failed to fetch the page.')
        return
    return json.loads(response.text)


if __name__ == '__main__':
    with open('the-trivia-api.csv', 'w') as file:
        csv_writer = csv.writer(file)
        headers = ['id', 'question', 'answer', 'alternatives',
                   'incorrect', 'source', 'tags', 'other']
        csv_writer.writerow(headers)
        i = 1
        seen = {}
        start = time.time()
        while True:
            questions = get_url_json(
                'https://the-trivia-api.com/v2/questions?limit=50')
            assert (len(questions) == 50)
            seen_count = 0
            for question in questions:
                qid = question['id']
                if qid in seen:
                    print(f'seen question, {seen_count}')
                    print()
                    seen_count += 1
                    seen[qid] += 1
                    continue
                seen[qid] = 1
                q = TheTriviaAPIQuestion(question)
                q.csv_output(csv_writer)
            print(i, time.strftime('%H:%M:%S', time.gmtime(time.time() - start)))
            print('Finished processing response')
            print()
            i += 1
            if seen_count == 50:
                print('stopping after no new questions found on request', i)
                break
