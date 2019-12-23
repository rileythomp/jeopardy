#include "helpers.h"
#include "title_print.h"
using namespace std;
using namespace pqxx;

int main(int argc, char* argv[]) {
    bool password_wrong = true;
    while (password_wrong) {
        try {
            // Initialize db
            // string pgpw = getenv("POSTGRES_PW");
            cout << "Enter database password:" << endl;
            string pgpw;
            getline(cin, pgpw);
            connection conn("dbname = testdb user = postgres password = " + pgpw + " hostaddr = 127.0.0.1 port = 5432");
            password_wrong = false;
            check_open(conn);
            prepare_jeopardy_queries(conn);
            print_title();	
            cout << endl;
            cout << "Enter response as 'skipq' to skip a question" << endl;
            cout << "Enter response as 'quitround' or 'quitgame' to end a round or game respectively" << endl << endl;

            loop_game(conn);

            conn.disconnect();
        }
        catch (const exception &e) {
            cout << "Sorry that password is incorrect, try again" << endl;
        }
    }
    return 0;
}
