#ifndef QUESTION_H
#define QUESTION_H

#include <string>
#include <iostream>
#include <sstream>
#include <iterator>
#include <vector>
#include <pqxx/pqxx>

class Question {
    public:
      Question(std::vector<std::string> data, int _season) {
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

      Question(pqxx::result::const_iterator data) {
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

      int get_value() {
        return value;
      }

      std::string get_response() {
        return question;
      }

      void ask() {
        std::cout << "Value: " << value << std::endl;
        std::cout << "Category: " << category << std::endl;
        if (comments != "-") {
          std::cout << "Comments: " << comments << std::endl;
        }
        std::cout << "Clue: " << answer << std::endl;
      }

      bool correct_response(std::string user_response) {
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

      friend std::ostream& operator<< (std::ostream& out, const Question& q) {
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

    private:
      int season;
      int round;
      int value;
      std::string daily_double;
      std::string category;
      std::string comments;
      std::string answer; // what is asked
      std::string question; // what is responded
      std::string air_date;
      std::string notes;

      void lower_case(std::string& str) {
        for (int i = 0; i < str.length(); ++i) {
          str[i] = tolower(str[i]);
        }
      }
};

#endif