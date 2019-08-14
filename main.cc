#include "question.h"

using namespace std;
using namespace pqxx;

// Convert string to upper case
void upper_case(string& str) {
  for (int i = 0; i < str.length(); ++i) {
    str[i] = toupper(str[i]);
  }
}

// Update question index
void update_index(bool& reset_index, int& i) {
  if (reset_index) {
    i = max(-1, i-10);
    reset_index = false;
  }
  ++i;
}

// Update score values
void update_vals(int& val1, int incr1, int& val2, int incr2) {
  val1 += incr1;
  val2 += incr2;
}

// Get user input
void get_input(string& str, string msg) {
  cout << msg << ": ";
  getline(cin, str);
}

// Get list of questions
void get_questions(result& questions, work& work, string category, string value) {
  if (category == "" && value == "") {
    questions = work.prepared("select_all_questions").exec();
  }
  else if (category == "") {
    questions = work.prepared("select_by_value")(stoi(value)).exec();
  }
  else if (value == "") {
    questions = work.prepared("select_by_category")("%"+category+"%").exec();
  }
  else {
    questions = work.prepared("select_by_category_and_value")("%"+category+"%")(stoi(value)).exec();
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

    // Prepare SQL queries
    conn.prepare("select_all_questions", "SELECT * FROM QUESTIONS WHERE DAILY_DOUBLE = 'no' AND VALUE != 0");
    conn.prepare("select_by_value", "SELECT * FROM QUESTIONS WHERE VALUE = $1 AND DAILY_DOUBLE = 'no'");
    conn.prepare("select_by_category", "SELECT * FROM QUESTIONS WHERE CATEGORY LIKE $1 AND DAILY_DOUBLE = 'no' AND VALUE != 0 ORDER BY CATEGORY DESC");
    conn.prepare("select_by_category_and_value", "SELECT * FROM QUESTIONS WHERE CATEGORY LIKE $1 AND VALUE = $2 AND DAILY_DOUBLE = 'no' ORDER BY CATEGORY DESC");

    // Intialize game variables
    srand(time(nullptr));
    int total_asked = 0;
    int total_correct = 0;
    int value_asked = 0;
    int value_correct = 0;
    bool quit_game = false;

    cout << "Jeopardy Practice" << endl << endl;
    cout << "Enter response as 'skipq' to skip a question" << endl;
    cout << "Enter response as 'quitround' or 'quitgame' to end a round or game respectively" << endl << endl;

    // Game loop
    while (!quit_game) {
      // Set category
      string category;
      get_input(category, "Category");
      upper_case(category);

      // Set value
      string value;
      get_input(value, "Value");

      // Get questions
      work work(conn);
      result questions;
      get_questions(questions, work, category, value);
      work.commit();

      int len = questions.size();

      // Handle no questions returned
      if (len == 0) {
        cout << "No clues matched your query, try again" << endl << endl;
        continue;
      }

      cout << "Total questions: " << len << endl << endl;

      // Initialize round variables
      int round_correct = 0;
      int round_asked = 0;
      int round_val_correct = 0;
      int round_val_asked = 0;
      int max_before_reset = len-1;
      int i = max(0, len-5); 
      bool quit_round = false;
      bool reset_index = false;

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
        string response;
        get_input(response, "Response");

        // Handle quit responses
        if (response == "quitround") {
          break;
        }
        else if (response == "quitgame") {
          quit_game = true;
          break;
        }
        else if (response == "skipq") {
          update_index(reset_index, i);
          cout << "Full response: " << q.get_response() << endl << endl;
          continue;
        }

        // Handle response
        if (q.is_correct_response(response)) {
          cout << "Correct!" << endl;
          update_vals(round_correct, 1, round_val_correct, q.get_value());
        }
        else {
          cout << "Incorrect" << endl;
        }

        // Update round totals
        update_vals(round_asked, 1, round_val_asked, q.get_value());

        cout << "Full response: " << q.get_response() << endl;
        cout << "Round Score: " << round_correct << "/" << round_asked << ", Round Value: " << round_val_correct << "/" << round_val_asked << endl;
        cout << endl;

        // Go down the list 10 questions, to start of new 5-set
        update_index(reset_index, i);
      }

      // Update game variables
      update_vals(total_correct, round_correct, total_asked, round_asked);
      update_vals(value_correct, round_val_correct, value_asked, round_val_asked);

      cout << "Total Score: " << total_correct << "/" << total_asked << ", Total Value: " << value_correct << "/" << value_asked << endl << endl;
    }

    conn.disconnect();
  }
  catch (const exception &e) {
    cerr << e.what() << endl;
    return 1;
  }

  return 0;
}
