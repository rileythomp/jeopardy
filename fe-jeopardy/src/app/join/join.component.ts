import { Component } from '@angular/core'
import { Router } from '@angular/router'
import { JwtService } from '../services/jwt.service'
import { ApiService } from '../services/api.service'
import { Observer } from 'rxjs'
import { ServerUnavailableMsg } from '../model/model'
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
	protected botConfig: number = 0
	protected pickConfig: number = 30
	protected buzzConfig: number = 30
	protected answerConfig: number = 15
	protected voteConfig: number = 10
	protected wagerConfig: number = 30
	protected firstRoundCategories: any[] = []
	protected secondRoundCategories: any[] = []


	constructor(
		private router: Router,
		private jwtService: JwtService,
		private apiService: ApiService,
		protected modal: ModalService,
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

	createPrivateGame(bots: number) {
		this.apiService.CreatePrivateGame(
			this.playerName,
			bots, this.twoRoundChecked, this.penaltyChecked,
			this.pickConfig, this.buzzConfig, this.answerConfig, this.voteConfig, this.wagerConfig,
			this.firstRoundCategories, this.secondRoundCategories
		).subscribe(this.joinResp())
	}

	joinGameByCode(gameCode: string) {
		this.apiService.JoinGameByCode(this.playerName, gameCode).subscribe(this.joinResp())
	}

	joinPublicGame() {
		this.apiService.JoinPublicGame(this.playerName).subscribe(this.joinResp())
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
			this.joinGameByCode(this.gameCode)
			return
		}
		this.showGameCodeInput = !this.showGameCodeInput
	}
}
