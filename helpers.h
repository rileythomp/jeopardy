#ifndef HELPERS_H
#define HELPERS_H

#include "database.h"
#include "question.h"

void normalize_string(std::string& str, int (*change_case)(int c)) {
  for (int i = 0; i < str.length(); ++i) {
    str[i] = (*change_case)(str[i]);
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

void loop_round(
  const int& len,
  pqxx::result& questions,
  bool& quit_game,
  int& round_asked,
  int& round_correct,
  int& round_val_asked,
  int& round_val_correct) {
  // Round loop
  for (auto question : questions) {
    // Construct and ask question
    Question q = Question(question);
    q.ask();

    // Get response
    std::string response;
    get_input(response, "Response");
    normalize_string(response, tolower);

    // Handle quit responses
    if (response == "quitround") {
      q.end(round_correct, round_asked, round_val_correct, round_val_asked);
      break;
    }
    else if (response == "quitgame") {
      quit_game = true;
      q.end(round_correct, round_asked, round_val_correct, round_val_asked);
      break;
    }
    else if (response == "skipq") {
      q.end(round_correct, round_asked, round_val_correct, round_val_asked);
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

    q.end(round_correct, round_asked, round_val_correct, round_val_asked);
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
    normalize_string(category, toupper);

    // Set value
    bool invalid_value = true;
    int value;
    while (invalid_value) {
      std::string value_input;
      get_input(value_input, "Value");
      try {
        value = stoi(value_input);
        invalid_value = false;
      } catch (const std::exception &e) {
        std::cout << "Sorry, that's not a valid value, pelase enter a number or just hit enter" << std::endl;
      }
    }

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
    std::cout << "Thanks for playing" << std::endl << std::endl;
  }
}

#endif
