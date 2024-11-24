import ast
import csv
import html
import json
import sys
import time
import uuid

import requests


class OpenTriviaDBQuestion:
    def __init__(self, question):
        self.id = uuid.uuid4()
        self.question = question['question']
        self.answer = question['correct_answer']
        self.alternatives = []
        self.source = 'opentdb.com'
        self.tags = [question['category'].lower()]
        self.other = {
            'difficulty': question['difficulty'],
        }
        self.incorrect = question['incorrect_answers']

    def output(self, output=sys.stdout):
        print(self.id, self.question, self.answer, self.alternatives,
              self.incorrect, self.source, self.tags, self.other, file=output)

    def csv_output(self, writer):
        writer.writerow([self.id, self.question, self.answer, self.alternatives,
                        self.incorrect, self.source, self.tags, self.other])


def get_url_json(url: str):
    response = requests.get(url)
    if response.status_code != 200:
        print(response)
        print('Failed to fetch the page.')
        return
    return json.loads(response.text)


if __name__ == '__main__':
    # rows = []
    # processed = 0
    # with open('opentdb.csv', 'r') as file:
    #     csv_reader = csv.reader(file)
    #     for row in csv_reader:
    #         processed += 1
    #         if processed == 1:
    #             continue
    #         print(row)
    #         row[1] = html.unescape(row[1])
    #         row[2] = html.unescape(row[2])
    #         row[4] = [html.unescape(incorrect)
    #                   for incorrect in ast.literal_eval(row[4])]
    #         row[6] = [html.unescape(tag) for tag in ast.literal_eval(row[6])]
    #         rows.append(row)

    #     with open('tmp.csv', 'w') as file:
    #         csv_writer = csv.writer(file)
    #         csv_writer.writerow(['id', 'question', 'answer', 'alternatives',
    #                              'incorrect', 'source', 'tags', 'other'])
    #         for row in rows:
    #             csv_writer.writerow(row)
    # exit()

    with open('opentdb.csv', 'w') as file:
        csv_writer = csv.writer(file)
        headers = ['id', 'question', 'answer', 'alternatives',
                   'incorrect', 'source', 'tags', 'other']
        csv_writer.writerow(headers)
        i = 1
        seen = {}
        start = time.time()
        while True:
            resp = get_url_json(
                'https://opentdb.com/api.php?amount=50&type=multiple')
            questions = resp['results']
            assert (len(questions) == 50)
            seen_count = 0
            for question in questions:
                qid = question['question']
                if qid in seen:
                    seen_count += 1
                    seen[qid] += 1
                    print(f'seen question, {seen_count}')
                    print()
                    continue
                seen[qid] = 1
                q = OpenTriviaDBQuestion(question)
                q.csv_output(csv_writer)
            print(i, time.strftime('%H:%M:%S', time.gmtime(time.time() - start)))
            print('Finished processing response')
            print()
            i += 1
            if seen_count == 50:
                print('stopping after no new questions found on request', i)
                break
            time.sleep(5)
