import { Component } from '@angular/core'
import { Router } from '@angular/router'
import { Observer } from 'rxjs'
import { ServerUnavailableMsg } from '../model/model'
import { ApiService } from '../services/api.service'
import { AuthService } from '../services/auth.service'
import { JwtService } from '../services/jwt.service'
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
	protected oneRoundChecked: boolean = true
	protected twoRoundChecked: boolean = false
	protected penaltyChecked: boolean = true

	constructor(
		private router: Router,
		private jwtService: JwtService,
		private apiService: ApiService,
		protected modal: ModalService,
		protected user: AuthService,
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

	private showInvalidName() {
		document.getElementById('player-name-join')!.focus()
		document.getElementById('player-name-join')!.style.border = '1px solid red';
		setTimeout(() => {
			document.getElementById('player-name-join')!.style.border = '1px solid grey';
		}, 1000)
	}

	createPrivateGame(bots: number) {
		let playerImg = ''
		if (this.user.Authenticated()) {
			this.playerName = this.user.Name()
			playerImg = this.user.ImgUrl()
		}
		if (this.playerName == '') {
			this.showInvalidName()
			return
		}
		this.apiService.CreatePrivateGame(
			this.playerName, playerImg,
			bots,
			this.twoRoundChecked,
			this.penaltyChecked,
			30, 30, 15, 30, [], []
		).subscribe(this.joinResp())
	}

	joinGameByCode() {
		let playerImg = ''
		if (this.user.Authenticated()) {
			this.playerName = this.user.Name()
			playerImg = this.user.ImgUrl()
		}
		if (this.playerName == '') {
			this.showInvalidName()
			return
		}
		this.apiService.JoinGameByCode(this.playerName, playerImg, this.gameCode).subscribe(this.joinResp())
	}

	joinPublicGame() {
		let playerImg = ''
		if (this.user.Authenticated()) {
			this.playerName = this.user.Name()
			playerImg = this.user.ImgUrl()
		}
		if (this.playerName == '') {
			this.showInvalidName()
			return
		}
		this.apiService.JoinPublicGame(this.playerName, playerImg).subscribe(this.joinResp())
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
		if (this.user.Authenticated()) {
			this.playerName = this.user.Name()
		}
		if (this.showGameCodeInput && this.gameCode && this.playerName) {
			this.joinGameByCode()
			return
		}
		this.showGameCodeInput = !this.showGameCodeInput
	}
}
