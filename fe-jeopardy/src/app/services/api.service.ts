import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { Observable } from 'rxjs';

const apiAddr = environment.apiServerUrl;
const httpProtocol = environment.httpProtocol;

@Injectable({
    providedIn: 'root'
})
export class ApiService {

    constructor(private http: HttpClient) { }

    CreatePrivateGame(playName: string, bots: number): Observable<any> {
        return this.post('games', {
            playerName: playName,
            bots: bots,
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
        })
    }

    AddBot(): Observable<any> {
        return this.put(`games/bot`, {})
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
            }),
            withCredentials: true,
        }

    }
}
