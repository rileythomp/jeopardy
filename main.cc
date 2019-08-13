#include <fstream>
#include "question.h"

using namespace std;
using namespace pqxx;

void upper_case(string& str) {
  for (int i = 0; i < str.length(); ++i) {
    str[i] = toupper(str[i]);
  }
}

int main(int argc, char* argv[]) {
  try {
    // Initialize db
    string pgpw = getenv("POSTGRES_PW");
    connection conn("dbname = testdb user = postgres password = " + pgpw + " hostaddr = 127.0.0.1 port = 5432");
    if (!conn.is_open()) {
      cout << "Can't open database" << endl;
      return 1;
    }

    // Intialize game variables
    srand(time(nullptr));
    int total_asked, total_correct, value_asked, value_correct = 0;
    bool quit_game = false;

    cout << "Jeopardy Practice" << endl << endl;
    cout << "Enter response as 'quitround' or 'quitgame' to end a round or game respectively" << endl << endl;

    // Game loop
    while (!quit_game) {
      // Set category
      cout << "Category: ";
      string category;
      getline(cin, category);
      upper_case(category);

      // Set value
      cout << "Value: ";
      string value;
      getline(cin, value);

      // Get questions
      work work(conn);
      result questions;
      if (category == "" && value == "") { // both empty
        conn.prepare("select_all_questions", "SELECT * FROM QUESTIONS WHERE DAILY_DOUBLE = $1 AND VALUE != 0");
        questions = work.prepared("select_all_questions")("no").exec();
      }
      else if (category == "") { // category empty, value set
        conn.prepare("select_by_value", "SELECT * FROM QUESTIONS WHERE VALUE = $1 AND DAILY_DOUBLE = $2");
        questions = work.prepared("select_by_value")(stoi(value))("no").exec();
      }
      else if (value == "") { // category set, value empty
        conn.prepare("select_by_category", "SELECT * FROM QUESTIONS WHERE CATEGORY LIKE $1 AND DAILY_DOUBLE = $2 AND VALUE != 0 ORDER BY CATEGORY DESC");
        questions = work.prepared("select_by_category")("%"+category+"%")("no").exec();
      }
      else { // both set
        conn.prepare("select_by_category_and_value", "SELECT * FROM QUESTIONS WHERE CATEGORY = $1 AND VALUE = $2 AND DAILY_DOUBLE = $3 ORDER BY CATEGORY DESC");
        questions = work.prepared("select_by_category_and_value")("%"+category+"%")(stoi(value))("no").exec();
      }
      work.commit();

      int len = questions.size();

      // Handle no questions returned
      if (len == 0) {
        cout << "No clues matched your query, try again" << endl << endl;
        continue;
      }

      cout << "Total questions: " << len << endl << endl;

      // Initialize round variables
      int round_correct, round_asked, round_val_correct, round_val_asked = 0;
      int max_before_reset = len-1;
      int i = max(0, len-5); 
      bool quit_round, reset_index = false;

      // Round loop
      while (!quit_round) {
        // Handle end of 5-set
        if (i == max_before_reset) {
          quit_round = max_before_reset < 5;
          reset_index = true;
          max_before_reset -= 5;
        }

        // Construct and ask question
        Question q = Question(questions[i]);
        q.ask();

        // Get response
        std::cout << "Response: ";
        std::string response;
        std::getline(std::cin, response);

        // Handle quit responses
        if (response == "quitround") {
          break;
        }
        else if (response == "quitgame") {
          quit_game = true;
          break;
        }

        // Handle response
        if (q.is_correct_response(response)) {
          std::cout << "Correct!" << std::endl;
          round_correct += 1;
          round_val_correct += q.get_value();
        }
        else {
          std::cout << "Incorrect" << std::endl;
        }

        // Update round totals
        round_asked += 1;
        round_val_asked += q.get_value();

        std::cout << "Full response: " << q.get_response() << std::endl;
        cout << "Round Score: " << round_correct << "/" << round_asked << " Round Value: " << round_val_correct << "/" << round_val_asked << endl;
        std::cout << std::endl;

        // Go down the list 10 questions, to start of new 5-set
        if (reset_index) {
          i = max(-1, i-10);
          reset_index = false;
        }
        ++i;
      }

      // Update game variables
      total_correct += round_correct;
      total_asked += round_asked;
      value_correct += round_val_correct;
      value_asked += round_val_asked;

      cout << "Score: " << total_correct << "/" << total_asked << " Value: " << value_correct << "/" << value_asked << endl << endl;
    }

    conn.disconnect();
  }
  catch (const std::exception &e) {
    cerr << e.what() << std::endl;
    return 1;
  }

  return 0;
}
