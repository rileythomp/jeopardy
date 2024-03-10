import { Component } from '@angular/core'
import { Router } from '@angular/router'
import { JwtService } from '../services/jwt.service'
import { ApiService } from '../services/api.service'
import { Observer } from 'rxjs'
import { ServerUnavailableMsg } from '../constants'
import { ModalService } from '../services/modal.service'

@Component({
	selector: 'app-join',
	templateUrl: './join.component.html',
	styleUrls: ['./join.component.less'],
})
export class JoinComponent {
	protected playerName: string = ''
	protected gameCode: string = ''
	protected showGameCodeInput: boolean = false
	protected oneRoundChecked: boolean = false
	protected twoRoundChecked: boolean = true

	constructor(
		private router: Router,
		private jwtService: JwtService,
		private apiService: ApiService,
		private modal: ModalService,
	) { }

	private joinResp(): Partial<Observer<any>> {
		return {
			next: (resp: any) => {
				this.jwtService.SetJWT(resp.token)
				this.router.navigate([`/game/${resp.game.name}`])
			},
			error: (err: any) => {
				let msg = err.status != 0 ? err.error.message : ServerUnavailableMsg;
				this.modal.displayMessage(msg)
			}
		}
	}

	createPrivateGame(playerName: string, bots: number) {
		this.apiService.CreatePrivateGame(playerName, bots, this.twoRoundChecked).subscribe(this.joinResp())
	}

	joinGameByCode(playerName: string, gameCode: string) {
		this.apiService.JoinGameByCode(playerName, gameCode).subscribe(this.joinResp())
	}

	joinPublicGame(playerName: string) {
		this.apiService.JoinPublicGame(playerName).subscribe(this.joinResp())
	}

	rejoin() {
		this.apiService.GetPlayerGame().subscribe({
			next: (resp: any) => {
				this.router.navigate([`/game/${resp.game.name}`])
			},
			error: (err: any) => {
				let msg = err.status != 0 ? err.error.message : ServerUnavailableMsg
				this.modal.displayMessage(msg)
			},
		})
	}

	toggleGameCodeInput() {
		if (this.showGameCodeInput && this.gameCode && this.playerName) {
			this.joinGameByCode(this.playerName, this.gameCode)
			return
		}
		this.showGameCodeInput = !this.showGameCodeInput
	}
}
