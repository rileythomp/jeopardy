import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { Observable } from 'rxjs';
import { JwtService } from './jwt.service';

const apiAddr = environment.apiServerUrl;
const httpProtocol = environment.httpProtocol;

@Injectable({
    providedIn: 'root'
})
export class ApiService {

    constructor(
        private http: HttpClient,
        private jwtService: JwtService,
    ) { }

    CreatePrivateGame(
        playName: string, bots: number, fullGame: boolean, penalty: boolean,
        pickConfig: number, buzzConfig: number, answerConfig: number, voteConfig: number, wagerConfig: number,
        firstRoundCategories: any[], secondRoundCategories: any[]
    ): Observable<any> {
        return this.post('games', {
            playerName: playName, bots: bots, fullGame: fullGame, penalty: penalty,
            pickConfig: pickConfig, buzzConfig: buzzConfig, answerConfig: answerConfig, voteConfig: voteConfig, wagerConfig: wagerConfig,
            firstRoundCategories: firstRoundCategories, secondRoundCategories: secondRoundCategories,
        })
    }

    JoinGameByCode(playerName: string, gameCode: string): Observable<any> {
        return this.put(`games/${gameCode}`, {
            playerName: playerName,
            gameCode: gameCode,
        })
    }

    JoinPublicGame(playerName: string): Observable<any> {
        return this.put('games', {
            playerName: playerName,
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
                'Access-Token': this.jwtService.GetJWT(),
            })
        }

    }
}
