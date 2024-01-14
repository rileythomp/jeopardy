import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { Observable } from 'rxjs';
import { JwtService } from './jwt.service';

const apiAddr = environment.apiServerUrl;
const httpProtocol = environment.httpProtocol;

const JsonOpts = {
    headers: new HttpHeaders({
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    })
}

@Injectable({
    providedIn: 'root'
})
export class ApiService {

    constructor(
        private http: HttpClient,
        private jwtService: JwtService,
    ) { }

    CreatePrivateGame(playName: string): Observable<any> {
        return this.post('games', {
            playerName: playName,
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

    GetPlayerGame(): Observable<any> {
        return this.get('players/game')
    }

    LeaveGame(user: any): Observable<any> {
        return this.post('leave', user)
    }

    PlayAgain(user: any): Observable<any> {
        return this.post('play-again', user)
    }

    private post(path: string, req: any): Observable<any> {
        return this.http.post<any>(
            `${httpProtocol}://${apiAddr}/jeopardy/${path}`,
            req,
            JsonOpts
        )
    }

    private put(path: string, req: any): Observable<any> {
        return this.http.put<any>(
            `${httpProtocol}://${apiAddr}/jeopardy/${path}`,
            req,
            JsonOpts
        )
    }

    private get(path: string): Observable<any> {
        return this.http.get<any>(
            `${httpProtocol}://${apiAddr}/jeopardy/${path}`,
            this.getOpts(),
        )
    }

    private getOpts() {
        return {
            headers: new HttpHeaders({
                'Content-Type': 'application/json',
                'Accept': 'application/json',
                'Access-Token': this.jwtService.GetJWT(),
            })
        }

    }
}
