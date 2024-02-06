import { Component } from '@angular/core';
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
export class JoinComponent {
	playerName: string = '';
	gameCode: string = '';

	constructor(
		private router: Router,
		private jwtService: JwtService,
		private apiService: ApiService,
	) { }


	private joinResp(): Partial<Observer<any>> {
		return {
			next: (resp: any) => {
				this.jwtService.SetJWT(resp.token);
				this.router.navigate([`/game/${resp.game.name}`]);
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
		this.apiService.GetPlayerGame().subscribe({
			next: (resp: any) => {
				this.router.navigate([`/game/${resp.game.name}`]);
			},
			error: (resp: any) => {
				// TODO: REPLACE WITH MODAL 
				alert(resp.error.message);
			},
		});
	}
}
