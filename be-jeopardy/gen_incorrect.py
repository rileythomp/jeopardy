import csv
import os

from openai import OpenAI

client = OpenAI(
    organization=os.environ.get('OPENAI_ORG'),
    project=os.environ.get('OPENAI_PROJECT'),
    api_key=os.environ.get('OPENAI_API_KEY'),
)

system_prompt = '''
We are building a multiple choice trivia game. 
We have a large number of questions and a correct answer for each question. 
We need to generate 3 incorrect but plausible answers to each question. 
You will be given a question in the form: `question: "question_text", answer: "correct_answer"`, 
and you need to respond with 3 incorrect but plausible answers to the question in the form `['incorrect_1', 'incorrect_2', 'incorrect_3'].
'''

with open('input.csv', 'r') as input_file, open('output.csv', 'w') as output_file:
    input_reader = csv.reader(input_file)
    output_writer = csv.writer(output_file)
    headers = ['id', 'question', 'answer', 'alternatives',
               'incorrect', 'source', 'tags', 'other']
    output_writer.writerow(headers)
    header = False
    for row in input_reader:
        if not header:
            header = True
            continue
        question, answer = row[1], row[2]
        chat_completion = client.chat.completions.create(
            model='gpt-4o-mini',
            messages=[
                {
                    'role': 'system',
                    'content': system_prompt
                },
                {'role': 'user', 'content': f'question: "{question}", answer: "{answer}"'}
            ]
        )
        row[4] = chat_completion.choices[0].message.content
        output_writer.writerow(row)

# question: Who painted the ceiling of the Sistine chapel?
# answer:   Michelangelo
# incorrect: { Raphael, Leonardo da Vinci, Donatello }
