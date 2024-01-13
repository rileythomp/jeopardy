import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { JwtService } from '../services/jwt.service';
import { ApiService } from '../services/api.service';
import { Observer } from 'rxjs';
// import { generateFakeWordByLength } from 'fakelish';

@Component({
	selector: 'app-join',
	templateUrl: './join.component.html',
	styleUrls: ['./join.component.less'],
})
export class JoinComponent implements OnInit {
	playerName: string = '';
	gameCode: string = '';

	constructor(
		private router: Router,
		private jwtService: JwtService,
		private apiService: ApiService,
	) { }

	ngOnInit(): void {
		// (async ()=>{
		// 	this.playerName = await generateFakeWordByLength(7);
		// 	this.gameName = 'autojoined'
		// 	this.joinGame(true);
		// })();
	}

	private joinResp(): Partial<Observer<any>> {
		return {
			next: (resp: any) => {
				this.jwtService.SetJWT(resp.token);
				this.router.navigate(['/game']);
			},
			error: (resp: any) => {
				// TODO: REPLACE WITH MODAL
				alert(resp.error.message);
			}

		}
	}

	createPrivateGame(playerName: string) {
		this.apiService.CreatePrivateGame(playerName).subscribe(this.joinResp());
	}

	joinGameByCode(playerName: string, gameCode: string) {
		this.apiService.JoinGameByCode(playerName, gameCode).subscribe(this.joinResp());
	}

	joinPublicGame(playerName: string) {
		this.apiService.JoinPublicGame(playerName).subscribe(this.joinResp());
	}

	rejoin() {
		this.router.navigate(['/game']);
	}
}
