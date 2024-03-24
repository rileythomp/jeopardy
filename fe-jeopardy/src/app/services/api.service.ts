import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../environments/environment';
import { JwtService } from './jwt.service';

const apiAddr = environment.apiServerUrl;
const httpProtocol = environment.httpProtocol;

@Injectable({
    providedIn: 'root'
})
export class ApiService {

    constructor(
        private http: HttpClient,
        private jwt: JwtService,
    ) { }

    GetPlayerAnalytics(email: string): Observable<any> {
        return this.get(`analytics/players?email=${email}`)
    }

    CreatePrivateGame(
        playName: string, playerImg: string, playerEmail: string, bots: number, fullGame: boolean, penalty: boolean,
        pickConfig: number, buzzConfig: number, answerConfig: number, wagerConfig: number,
        firstRoundCategories: any[], secondRoundCategories: any[]
    ): Observable<any> {
        return this.post('games', {
            playerName: playName, playerImg: playerImg, playerEmail: playerEmail, bots: bots, fullGame: fullGame, penalty: penalty,
            pickConfig: pickConfig, buzzConfig: buzzConfig, answerConfig: answerConfig, wagerConfig: wagerConfig,
            firstRoundCategories: firstRoundCategories, secondRoundCategories: secondRoundCategories,
        })
    }

    JoinGameByCode(playerName: string, playerImg: string, playerEmail: string, joinCode: string): Observable<any> {
        return this.put(`games/${joinCode}`, {
            playerName: playerName,
            playerImg: playerImg,
            playerEmail: playerEmail,
            joinCode: joinCode,
        })
    }

    JoinPublicGame(playerName: string, playerImg: string, playerEmail: string): Observable<any> {
        return this.put('games', {
            playerName: playerName,
            playerImg: playerImg,
            playerEmail: playerEmail,
            fullGame: true,
        })
    }

    AddBot(): Observable<any> {
        return this.put('games/bot', {})
    }

    GetPlayerGame(): Observable<any> {
        return this.get('players/game')
    }

    LeaveGame(): Observable<any> {
        return this.post('leave', {})
    }

    PlayAgain(): Observable<any> {
        return this.put('play-again', {})
    }

    GetAnalytics(): Observable<any> {
        return this.get('analytics')
    }

    SearchCategories(category: string, bothRounds: boolean): Observable<any> {
        let rounds = bothRounds ? 'both' : 'first'
        return this.get(`categories?category=${category}&rounds=${rounds}`)
    }

    StartGame(): Observable<any> {
        return this.put('games/start', {})
    }

    private post(path: string, req: any): Observable<any> {
        return this.http.post<any>(
            `${httpProtocol}://${apiAddr}/jeopardy/${path}`,
            req,
            this.headers()
        )
    }

    private put(path: string, req: any): Observable<any> {
        return this.http.put<any>(
            `${httpProtocol}://${apiAddr}/jeopardy/${path}`,
            req,
            this.headers()
        )
    }

    private get(path: string): Observable<any> {
        return this.http.get<any>(
            `${httpProtocol}://${apiAddr}/jeopardy/${path}`,
            this.headers(),
        )
    }

    private headers() {
        return {
            headers: new HttpHeaders({
                'Content-Type': 'application/json',
                'Accept': 'application/json',
                'Access-Token': this.jwt.GetJWT(),
            })
        }

    }
}
