#include "question.h"

// Construct question from tsv data
Question::Question(std::vector<std::string> data, int _season) {
  season = _season;
  round = std::stoi(data[0]);
  value = std::stoi(data[1]);
  daily_double = data[2];
  category = data[3];
  comments = data[4];
  answer = data[5];
  question = data[6];
  air_date = data[7];
  notes = data[8];
}

// Construct question from postgres response
Question::Question(pqxx::result::const_iterator data) {
  round = data[0].as<int>();
  value = data[1].as<int>();
  daily_double = data[2].as<std::string>();
  category = data[3].as<std::string>();
  comments = data[4].as<std::string>();
  answer = data[5].as<std::string>();
  question = data[6].as<std::string>();
  air_date = data[7].as<std::string>();
  notes = data[8].as<std::string>();
  season = data[9].as<int>();
}

// Asks the user the question
void Question::ask() {
  std::cout << "Category: " << category << std::endl;
  std::cout << "Value:    " << value << std::endl;

  if (comments != "-") {
    std::cout << "Comments: " << comments << std::endl;
  }

  std::cout << "Clue:     " << answer << std::endl;
}

// convert string to lower case
void lower_case(std::string& str) {
  for (int i = 0; i < str.length(); ++i) {
    str[i] = tolower(str[i]);
  }
}

// determine if a word is a substring of another word
bool is_valid_substring(std::string s1, std::string s2) {
  return s2.length() > 2 && s1.find(s2) != std::string::npos;
}

// Determines if a response is correct
bool Question::is_correct_response(std::string user_response) {
  // Handle user entering nothing
  if (user_response == "") {
    return false;
  }
  
  // Separate answer into words
  std::istringstream iss(question);
  std::vector<std::string> words{std::istream_iterator<std::string>{iss},
                                 std::istream_iterator<std::string>{}};
  lower_case(user_response);

  for (auto word : words) {
    lower_case(word);

    if (is_valid_substring(word, user_response) ||
        is_valid_substring(user_response, word)) {
      return true;
    }
  }

  return user_response == question;
}

// Pretty prints a question
std::ostream& operator<< (std::ostream& out, const Question& q) {
  out << "Round: " << q.round << std::endl;
  out << "Value: " << q.value << std::endl;
  out << "Daily Double: " << q.daily_double << std::endl;;
  out << "Category: " << q.category << std::endl;
  out << "Comments: " << q.comments << std::endl;
  out << "Answer: " << q.answer << std::endl;
  out << "Question: " << q.question << std::endl;
  out << "Air date: " << q.air_date << std::endl;
  out << "Notes: " << q.notes << std::endl;
  out << "Season: " << q.season << std::endl;
  out << std::endl;
  
  return out;
}