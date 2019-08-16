#include <string>
#include <vector>
#include <iostream>

using std::getline;
using std::cin;
using std::cout;
using std::endl;
using std::vector;
using std::string;

vector<string> a = {"        ",
                    "        ",
                    " ____   ",
                    "|___  \\ ",
                    "/ __  | ",
                    "\\_____/ ",
                    "        ",
                    "        "};

vector<string> b = {"  _     ",
                    "| |     ",
                    "| |__   ",
                    "|  _  \\ ",
                    "| |_) | ",
                    "|____/  ",
                    "        ",
                    "        ",
                    "        "};

vector<string> c = {"        ",
                    "        ",
                    "  ____  ",
                    " /  __| ",
                    "|  (__  ",
                    " \\____| ",
                    "        ",
                    "        "};

vector<string> d = {"     _  ",
                    "    | | ",
                    "  __| | ",
                    " / _  | ",
                    "| (_| | ",
                    " \\____| ",
                    "        ",
                    "        ",
                    "        "};

vector<string> e = {"        ",
                    "        ",
                    "  ____  ",
                    " / __ \\ ",
                    "|  ___/ ",
                    " \\____| ",
                    "        ",
                    "        "};

vector<string> f = {" ___  ",
                    "|  _| ",
                    "| |_  ",
                    "|  _| ",
                    "| |   ",
                    "|_|   ",
                    "      ",
                    "      ",
                    "      "};

vector<string> g = {"        ",
                    "        ",
                    "  ____  ",
                    " / _  \\ ",
                    "| (_) | ",
                    " \\__  | ",
                    "  __| | ",
                    " \\____/ "};  

vector<string> h = {" _     ",
                    "| |    ",
                    "| |_   ",
                    "|    \\ ",
                    "| || | ",
                    "|_||_| ",
                    "       ",
                    "       ",
                    "       "};                    

vector<string> i = {" _  ",
                    "|_| ",
                    " _  ",
                    "| | ",
                    "| | ",
                    "|_| ",
                    "    ",
                    "    "};

vector<string> j = {"  _  ",
                    " |_| ",
                    "  _  ",
                    " | | ",
                    " | | ",
                    " | | ",
                    "_| | ",
                    "\\ _| "};

vector<string> k = {" _      ",
                    "| | __  ",
                    "| |/ /  ",
                    "|   \\   ",
                    "| |\\ \\  ",
                    "|_| \\_\\ ",
                    "        ",
                    "        "};
    
vector<string> l = {" _  ",
                    "| | ",
                    "| | ",
                    "| | ",
                    "| | ",
                    "|_| ",
                    "    ",
                    "    "};
vector<string> m = {"        ",
                    "        ",
                    " __  _  ",
                    "|  \\/ | ",
                    "| ||| | ",
                    "|_|||_| ",
                    "        ",
                    "        "};

vector<string> n = {"       ",
                    "       ",
                    " ___   ",
                    "|   \\  ",
                    "| || | ",
                    "|_||_| ",
                    "       ",
                    "       "};

vector<string> o = {"        ",
                    "        ",
                    "  ___   ",
                    " / _ \\  ",
                    "| (_) | ",
                    " \\___/  ",
                    "        ",
                    "        "};

vector<string> p = {"        ",
                    "        ",
                    " ____   ",
                    "|  _ \\  ",
                    "| (_) | ",
                    "|  __/  ",
                    "| |     ",
                    "|_|     "};

vector<string> q = {"        ",
                    "        ",
                    "  ____  ",
                    " / _  | ",
                    "| (_) | ",
                    " \\__  | ",
                    "    | | ",
                    "    |_/ "};
                    
vector<string> r = {"        ",
                    "        ",
                    " _   _  ",
                    "| |/ _| ",
                    "|  /    ",
                    "|_|     ",
                    "        ",
                    "        "};

vector<string> s = {"        ",
                    "        ",
                    " ______ ",
                    "|  ___/ ",
                    " \\___ \\ ",
                    "/_____| ",
                    "        ",
                    "        "};

vector<string> t = {"   _    ",
                    " _| |_  ",
                    "|_   _| ",
                    "  | |   ",
                    "  | |   ",
                    "  |_|   ",
                    "        ",
                    "        "};

vector<string> u = {"       ",
                    "       ",
                    " _  _  ",
                    "| || | ",
                    "| || | ",
                    " \\__/  ",
                    "       ",
                    "       "};

vector<string> v = {"       ",
                    "       ",
                    "_    _ ",
                    "\\ \\/ / ",
                    " \\  /  ",
                    "  \\/   ",
                    "       ",
                    "       "};

vector<string> w = {"         ",
                    "         ",
                    " _     _ ",
                    "\\ \\/\\/ / ",
                    " \\    /  ",
                    "  \\/\\/   ",
                    "         ",
                    "         "};

vector<string> x = {"        ",
                    "        ",
                    "_    _ ",
                    "\\ \\/ /",
                    " |  |  ",
                    "/_/\\_\\ ",
                    "        ",
                    "        "};

vector<string> y = {"        ",
                    "        ",
                    " _    _ ",
                    " \\ \\/ / ",
                    "  \\  /  ",
                    "  / /   ",
                    " / /    ",
                    "/_/     "};

vector<string> z = {"        ",
                    "        ",
                    "______  ",
                    "\\___  | ",
                    " / ___/ ",
                    "|_____\\ ",
                    "        ",
                    "        "};

vector<string> A = {"      **      ",
                    "     *  *     ",
                    "    *    *    ",
                    "   * **** *   ",
                    "  * *    * *  ",
                    " ***      *** "};

vector<string> B = {" *******  ",
                    " *  *   * ",
                    " *     *  ",
                    " *     *  ",
                    " *  *   * ",
                    " *******  "};

vector<string> C = {"   ******* ",
                    "  * ****** ",
                    " * *       ",
                    " * *       ",
                    "  * ****** ",
                    "   ******* "};

vector<string> D = {" *******   ",
                    " *      *  ",
                    " *   *   * ",
                    " *   *   * ",
                    " *      *  ",
                    " *******   "};

vector<string> E = {" ******* ",
                    " * *     ",
                    " * ***** ",
                    " * ***** ",
                    " * *     ",
                    " ******* "};

vector<string> F = {" ******** ",
                    " * *****  ",
                    " * *      ",
                    " * *****  ",
                    " * *      ",
                    " ***      "};

vector<string> G = {"   ********  ",
                    "  * *******  ",
                    " * *         ",
                    " * *    **** ",
                    "  * ***** *  ",
                    "   ********  "};

vector<string> H = {" ***   *** ",
                    " * *   * * ",
                    " * ***** * ",
                    " * ***** * ",
                    " * *   * * ",
                    " ***   *** "};

vector<string> I = {" *** ",
                    " * * ",
                    " * * ",
                    " * * ",
                    " * * ",
                    " *** "};

vector<string> J = {"       *** ",
		            "       * * ",
		            "       * * ",
 		            " * *   * * ",
		            "  * ***  * ",
		            "   ******  ",};

vector<string> K = {" ***  *** ",
		            " * * * *  ",
		            " * ** *   ",
		            " * ** *   ",
		            " * * * *  ",
		            " ***  *** "};

vector<string> L = {" ***      ",
                    " * *      ",
                    " * *      ",
                    " * *      ",
                    " * ****** ",
                    " ******** "};

vector<string> M = {" ***    *** ",
                    " *  *  *  * ",
                    " *   **   * ",
                    " *  ****  * ",
                    " * *    * * ",
                    " ***    *** "};

vector<string> N = {" ***   *** ",
                    " *  *  * * ",
                    " *   * * * ",
                    " * *  ** * ",
                    " * * *   * ",
                    " ***  **** "};

vector<string> O = {"    *****    ",
                    "  * ***** *  ",
                    " * *     * * ",
                    " * *     * * ",
                    "  * ***** *  ",
                    "    *****    "};

vector<string> P = {" ******  ",
                    " *  *  * ",
                    " *  *  * ",
                    " * ****  ",
                    " * *     ",
                    " ***     "};

vector<string> Q = {"    *****     ",
                    "  * ***** *   ",
                    " * *     * *  ",
                    " * *   * * *  ",
                    "  * ***** *   ",
                    "    *****   * "};

vector<string> R = {" ******   ",
                    " *  *  *  ",
                    " *  *  *  ",
                    " * *  *   ",
                    " * * * *  ",
                    " ***  *** "};

vector<string> S = {"  ****** ",
                    " * ***** ",
                    " * *     ",
                    "   * *   ",
                    " ***** * ",
                    " ******  "};

vector<string> T = {" ********* ",
                    " **** **** ",
                    "    * *    ",
                    "    * *    ",
                    "    * *    ",
                    "    * *    "};

vector<string> U = {" * *   * * ",
                    " * *   * * ",
                    " * *   * * ",
                    " * *   * * ",
                    " *  ***  * ",
                    "   *****   "};

vector<string> V = {" ***      *** ",
                    "  * *    * *  ",
                    "   * *  * *   ",
                    "    * ** *    ",
                    "     *  *     ",
                    "      **      "};

vector<string> W = {" ***           *** ",
                    "  * *         * *  ",
                    "   * *   *   * *   ",
                    "    * **   ** *    ",
                    "     *  ***  *     ",
                    "      **   **      "};

vector<string> X = {" ***   *** ",
                    "  * * * *  ",
                    "   * * *   ",
                    "   * * *   ",
                    "  * * * *  ",
                    " ***   *** "};

vector<string> Y = {" ***   *** ",
                    "  * * * *  ",
                    "   * * *   ",
                    "    * *    ",
                    "    * *    ",
                    "    ***    "};
vector<string> Z = {" ********* ",
                    " ******* * ",
                    "     * *   ",
                    "   * *     ",
                    " * ******* ",
                    " ********* "};

vector<string> space = {"      ",
                        "      ",
                        "      ",
                        "      ",
                        "      ",
                        "      ",
                        "      ",
                        "      ",
                        "      "};

vector<vector<string>> letters;
                    
void addtoletters(char letter) {
    switch (letter) {
        case 'a':
            letters.push_back(a);
            break;
        case 'b':
            letters.push_back(b);
            break;
        case 'c':
            letters.push_back(c);
            break;
        case 'd':
            letters.push_back(d);
            break;
        case 'e':
            letters.push_back(e);
            break;
        case 'f':
            letters.push_back(f);
            break;
        case 'g':
            letters.push_back(g);
            break;
        case 'h':
            letters.push_back(h);
            break;
        case 'i':
            letters.push_back(i);
            break;
        case 'j':
            letters.push_back(j);
            break;
        case 'k':
            letters.push_back(k);
            break;
        case 'l':
            letters.push_back(l);
            break;
        case 'm':
            letters.push_back(m);
            break;
        case 'n':
            letters.push_back(n);
            break;
        case 'o':
            letters.push_back(o);
            break;
        case 'p':
            letters.push_back(p);
            break;
        case 'q':
            letters.push_back(q);
            break;
        case 'r':
            letters.push_back(r);
            break;
        case 's':
            letters.push_back(s);
            break;
       case 't':
            letters.push_back(t);
            break;
        case 'u':
            letters.push_back(u);
            break;
        case 'v':
            letters.push_back(v);
            break;
        case 'w':
            letters.push_back(w);
            break;
        case 'x':
            letters.push_back(x);
            break;
        case 'y':
            letters.push_back(y);
            break;
        case 'z':
            letters.push_back(z);
            break;
        case 'A':
            letters.push_back(A);
            break;
        case 'B':
            letters.push_back(B);
            break;
        case 'C':
            letters.push_back(C);
            break;
        case 'D':
            letters.push_back(D);
            break;
        case 'E':
            letters.push_back(E);
            break;
        case 'F':
            letters.push_back(F);
            break;
        case 'G':
            letters.push_back(G);
            break;
        case 'H':
            letters.push_back(H);
            break;
        case 'I':
            letters.push_back(I);
            break;
        case 'J':
            letters.push_back(J);
            break;
        case 'K':
            letters.push_back(K);
            break;
        case 'L':
            letters.push_back(L);
            break;
        case 'M':
            letters.push_back(M);
            break;
        case 'N':
            letters.push_back(N);
            break;
        case 'O':
            letters.push_back(O);
            break;
        case 'P':
            letters.push_back(P);
            break;
        case 'Q':
            letters.push_back(Q);
            break;
        case 'R':
            letters.push_back(R);
            break;
        case 'S':
            letters.push_back(S);
            break;
       case 'T':
            letters.push_back(T);
            break;
        case 'U':
            letters.push_back(U);
            break;
        case 'V':
            letters.push_back(V);
            break;
        case 'W':
            letters.push_back(W);
            break;
        case 'X':
            letters.push_back(X);
            break;
        case 'Y':
            letters.push_back(Y);
            break;
        case 'Z':
            letters.push_back(Z);
            break;
        case ' ':
            letters.push_back(space);
            break;
    }
}

void print_title() {
	string word = "jeopardy practice";
    letters = {};
    for (char c : word) {
        addtoletters(c);
    }

    for (int x = 0; x < letters[0].size(); ++x) {
        for (auto letter : letters) {
            cout << letter[x];
        }
        cout << endl;
    }
}
