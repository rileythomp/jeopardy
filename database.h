#ifndef DATABASE_H
#define DATABASE_H

#include <pqxx/pqxx>
#include <iostream>

// Prepare SQL queries
void prepare_jeopardy_queries(pqxx::connection& conn) {
  conn.prepare(
    "select_all_questions", 
    "SELECT * FROM QUESTIONS WHERE DAILY_DOUBLE = 'no' AND VALUE != 0 ORDER BY AIR_DATE DESC, CATEGORY ASC, VALUE ASC");
  conn.prepare(
    "select_by_value",
    "SELECT * FROM QUESTIONS WHERE VALUE = $1 AND DAILY_DOUBLE = 'no' ORDER BY AIR_DATE DESC, CATEGORY ASC, VALUE ASC");
  conn.prepare(
    "select_by_category",
    "SELECT * FROM QUESTIONS WHERE CATEGORY LIKE $1 AND DAILY_DOUBLE = 'no' AND VALUE != 0 ORDER BY AIR_DATE DESC, CATEGORY ASC, VALUE ASC");
  conn.prepare(
    "select_by_category_and_value",
    "SELECT * FROM QUESTIONS WHERE CATEGORY LIKE $1 AND VALUE = $2 AND DAILY_DOUBLE = 'no' ORDER BY AIR_DATE DESC, CATEGORY ASC, VALUE ASC");
}

// Check if connection is open
void check_open(pqxx::connection& conn) {
  if (!conn.is_open()) {
    throw std::runtime_error("Can't open database");
  }
}

// Get list of questions
void get_questions(pqxx::result& questions, pqxx::work& work, std::string category, int value) {
  if (category == "" && value == 0) {
    questions = work.prepared("select_all_questions").exec();
  }
  else if (category == "") {
    questions = work.prepared("select_by_value")(value).exec();
  }
  else if (value == 0) {
    questions = work.prepared("select_by_category")("%"+category+"%").exec();
  }
  else {
    questions = work.prepared("select_by_category_and_value")("%"+category+"%")(value).exec();
  }
  work.commit();
}

#endif