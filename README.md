# jeopardy

![Screenshot from 2019-08-14 18-26-15](https://user-images.githubusercontent.com/35535783/63060790-3cd3f480-bec1-11e9-9abb-21a7f805f8c7.png)

A tool for practicing Jeopardy with over 300,000 real Jeopardy questions. Filter questions by category and/or difficulty. 

Built with:
 * C++
 * PostgreSQL

To run locally:

Must configure a PostgreSQL database to write questions to and have libpqxx installed.

```$ git clone https://github.com/rileythomp/jeopardy.git```

```$ cd jeopardy```

```$ g++ data_to_db.cc -o datatodb```

```$ ./datatodb```

```$ g++ main.cc question.cc -lpqxx -lpq -o jeopardy```

```$ ./jeopardy```

Note: Must compile with flags in that specific order.
