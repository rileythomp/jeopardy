#include "question.h"

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

void Question::ask() {
  std::cout << "Value: " << value << std::endl;
  std::cout << "Category: " << category << std::endl;

  if (comments != "-") {
    std::cout << "Comments: " << comments << std::endl;
  }

  std::cout << "Clue: " << answer << std::endl;
}

void lower_case(std::string& str) {
  for (int i = 0; i < str.length(); ++i) {
    str[i] = tolower(str[i]);
  }
}

bool Question::correct_response(std::string user_response) {
  if (user_response == "") {
    return false;
  }
  
  lower_case(user_response);
  std::istringstream iss(question);
  std::vector<std::string> words{std::istream_iterator<std::string>{iss}, std::istream_iterator<std::string>{}};

  for (auto word : words) {
    lower_case(word);

    if (user_response.find(word) != std::string::npos || word.find(user_response) != std::string::npos) {
      return true;
    }
  }

  return user_response == question;
}

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