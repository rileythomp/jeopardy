#include <iostream>
#include <sstream>
#include <fstream>
#include <vector>
#include <string>
#include <pqxx/pqxx> 
#include "question.h"

using namespace std;
using namespace pqxx;

void upper_case(string& str) {
  for (int i = 0; i < str.length(); ++i) {
    str[i] = toupper(str[i]);
  }
}

int main(int argc, char* argv[]) {
  srand(time(nullptr));
  try {
    connection C("dbname = testdb user = postgres password = password \
    hostaddr = 127.0.0.1 port = 5432");

    if (!C.is_open()) {
        cout << "Can't open database" << endl;
        return 1;
    }

    int total_asked = 0;
    int total_correct = 0;
    int value_asked = 0;
    int value_correct = 0;

    cout << "Jeopardy Practice" << endl;

    while (true) {
      cout << "Category: ";

      string category;

      getline(cin, category);

      upper_case(category);

      cout << "Value: ";

      string value;

      getline(cin, value);

      work W(C);

      result questions;

      if (category == "" && value == "") { // both empty
        C.prepare("select_all_questions", "SELECT * FROM QUESTIONS");
        questions = W.prepared("select_all_questions").exec();
      }
      else if (category == "") { // category empty, value set
        C.prepare("select_by_value", "SELECT * FROM QUESTIONS WHERE VALUE = $1");
        questions = W.prepared("select_by_value")(stoi(value)).exec();
      }
      else if (value == "") { // category set, value empty
        C.prepare("select_by_category", "SELECT * FROM QUESTIONS WHERE CATEGORY = $1");
        questions = W.prepared("select_by_category")(category).exec();
      }
      else { // both set
        C.prepare("select_by_category_and_value", "SELECT * FROM QUESTIONS WHERE CATEGORY = $1 AND VALUE = $2");
        questions = W.prepared("select_by_category_and_value")(category)(stoi(value)).exec();
      }

      W.commit();

      int len = questions.size();

      if (len == 0) {
        cout << "No clues matched your query, try again" << endl << endl;
        continue;
      }

      int round_correct = 0;
      int round_asked = 0;
      int round_val_correct = 0;
      int round_val_asked = 0;

      cout << "Total questions: " << len << endl << endl;
      for (int i = 0; i < len; ++i) {
        Question q = Question(questions[i]);

        q.ask();

        std::cout << "Response: ";
        std::string response;
        std::getline(std::cin, response);

        if (q.correct_response(response)) {
          std::cout << "Correct!" << std::endl;
          round_correct += 1;
          round_val_correct += q.get_value();
        }
        else {
          std::cout << "Incorrect" << std::endl;
        }
        round_asked += 1;
        round_val_asked += q.get_value();

        std::cout << "Full response: " << q.get_response() << std::endl;
        cout << "Round Score: " << round_correct << "/" << round_asked << " Round Value: " << round_val_correct << "/" << round_val_asked << endl;


        std::cout << std::endl;
      }
      total_correct += round_correct;
      total_asked += round_asked;
      value_correct += round_val_correct;
      value_asked += round_val_asked;
      cout << "Score: " << total_correct << "/" << total_asked << " Value: " << value_correct << "/" << value_asked << endl;

    }

    C.disconnect();
  }
  catch (const std::exception &e) {
    cerr << e.what() << std::endl;
    return 1;
  }

  return 0;
}
