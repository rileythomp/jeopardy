import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../environments/environment';
import { Observable } from 'rxjs';

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

    constructor(private http: HttpClient) { }

    JoinGame(playerName: string, gameName: string, privateGame: boolean): Observable<any> {
        return this.post('join', {
            playerName: playerName,
            gameName: gameName,
            private: privateGame,
        })
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
}
