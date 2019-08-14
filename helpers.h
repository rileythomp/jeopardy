#ifndef HELPERS_H
#define HELPERS_H

#include "database.h"
#include "question.h"

// Convert string to upper case
void upper_case(std::string& str) {
  for (int i = 0; i < str.length(); ++i) {
    str[i] = toupper(str[i]);
  }
}

// Update score values
void update_vals(int& val1, int incr1, int& val2, int incr2) {
  val1 += incr1;
  val2 += incr2;
}

// Get user input
void get_input(std::string& str, std::string msg) {
  std::cout << msg << ": ";
  std::getline(std::cin, str);
}

void loop_round(const int& len,
                pqxx::result& questions,
                bool& quit_game,
                int& round_asked,
                int& round_correct,
                int& round_val_asked,
                int& round_val_correct) {

  // Round loop
  for (int i = len-1; i >= 0; --i) {
    // Construct and ask question
    Question q = Question(questions[i]);
    q.ask();

    // Get response
    std::string response;
    get_input(response, "Response");

    // Handle quit responses
    if (response == "quitround") {
      std::cout << "Full response: " << q.get_response() << std::endl << std::endl;
      break;
    }
    else if (response == "quitgame") {
      quit_game = true;
      std::cout << "Full response: " << q.get_response() << std::endl << std::endl;
      break;
    }
    else if (response == "skipq") {
      std::cout << "Full response: " << q.get_response() << std::endl << std::endl;
      continue;
    }

    // Handle response
    if (q.is_correct_response(response)) {
      std::cout << "Correct!" << std::endl;
      update_vals(round_correct, 1, round_val_correct, q.get_value());
    }
    else {
      std::cout << "Incorrect" << std::endl;
    }

    // Update round totals
    update_vals(round_asked, 1, round_val_asked, q.get_value());

    std::cout << "Full response: " << q.get_response() << std::endl;
    std::cout << "Round Score: " << round_correct << "/" << round_asked << ", Round Value: " << round_val_correct << "/" << round_val_asked << std::endl;
    std::cout << std::endl;
  }
}

void loop_game(pqxx::connection& conn) {
  // Intialize game variables
  srand(time(nullptr));
  int total_asked = 0;
  int total_correct = 0;
  int value_asked = 0;
  int value_correct = 0;
  int round_correct = 0;
  int round_asked = 0;
  int round_val_correct = 0;
  int round_val_asked = 0;
  bool quit_game = false;

  // Game loop
  while (!quit_game) {
    // Set category
    std::string category;
    get_input(category, "Category");
    upper_case(category);

    // Set value
    std::string value;
    get_input(value, "Value");

    // Get questions
    pqxx::work work(conn);
    pqxx::result questions;
    get_questions(questions, work, category, value);
    // work.commit();

    const int len = questions.size();

    // Handle no questions returned
    if (len == 0) {
      std::cout << "No clues matched your query, try again" << std::endl << std::endl;
      continue;
    }

    std::cout << "Total questions: " << len << std::endl << std::endl;

    loop_round(len, questions, quit_game, round_asked, round_correct, round_val_asked, round_val_correct);

    // Update game variables
    update_vals(total_correct, round_correct, total_asked, round_asked);
    update_vals(value_correct, round_val_correct, value_asked, round_val_asked);

    std::cout << "Total Score: " << total_correct << "/" << total_asked << ", Total Value: " << value_correct << "/" << value_asked << std::endl << std::endl;
  }
}

#endif