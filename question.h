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
      // Construct question from tsv data
      Question(std::vector<std::string> data, int _season);

      // Construct question from postgres response
      Question(pqxx::result::const_iterator data);

      // Gets the question value
      inline int get_value() { return value; }

      // Gets the correct response
      inline std::string get_response() { return question; }

      // Asks the user the question
      void ask();

      // Prints response and round score
      void end(int rc, int ra, int rvc, int rva);

      // Determines if a response is correct
      bool is_correct_response(std::string user_response);

      // Pretty prints a question
      friend std::ostream& operator<< (std::ostream& out, const Question& q);

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
};

#endif