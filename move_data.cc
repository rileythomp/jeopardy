#include <iostream>
#include <sstream>
#include <fstream>
#include <vector>
#include <string>
#include <pqxx/pqxx> 
#include "question.h"

using namespace std;
using namespace pqxx;

int main(int argc, char* argv[]) {
  // Upload file data to PostgreSQL database

  connection C("dbname = testdb user = postgres password = password \
    hostaddr = 127.0.0.1 port = 5432");

  if (!C.is_open()) {
    cout << "Can't open database" << endl;
    return 1;
  }
  
  C.prepare("insert_to_questions", "INSERT INTO QUESTIONS (ROUND, VALUE, DAILY_DOUBLE, CATEGORY, COMMENTS, ANSWER, QUESTION, AIR_DATE, NOTES, SEASON) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);");
  
  const int num_seasons = 35;

  for (int i = 1; i <= num_seasons; ++i) {
    cout << "season " << i << "..." << endl;
    ifstream file("seasons/season" + to_string(i) + ".tsv");
    string line;

    getline(file, line); // skip the column headers
    while (getline(file, line)) {
      istringstream iss(line);
      vector<string> question;
      string question_data;

      while (getline(iss, question_data, '\t')) {
        question.emplace_back(question_data);
      }

      work W(C);
      W.prepared("insert_to_questions")(question[0])(question[1])(question[2])(question[3])(question[4])(question[5])(question[6])(question[7])(question[8])(i).exec();
      W.commit();
    }
  cout << "done" << endl;
  }
}
