import { Component, OnInit } from '@angular/core'
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
export class JoinComponent implements OnInit {
	protected playerName: string = ''
	private playerImg: string = ''
	protected userAuthenticated: boolean = false
	protected joinCode: string = ''
	protected showJoinCodeInput: boolean = false
	protected oneRoundChecked: boolean = true
	protected twoRoundChecked: boolean = false
	protected penaltyChecked: boolean = true

	constructor(
		private router: Router,
		private jwtService: JwtService,
		private apiService: ApiService,
		protected modal: ModalService,
		private auth: AuthService,
	) { }

	ngOnInit() {
		this.auth.user.subscribe(user => {
			this.userAuthenticated = user.authenticated
			this.playerName = user.name
			this.playerImg = user.imgUrl
		})
	}

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
		if (this.playerName == '') {
			this.showInvalidName()
			return
		}
		this.apiService.CreatePrivateGame(
			this.playerName, this.playerImg,
			bots, this.twoRoundChecked, this.penaltyChecked,
			30, 30, 15, 30, [], []
		).subscribe(this.joinResp())
	}

	joinGameByCode() {
		if (this.playerName == '') {
			this.showInvalidName()
			return
		}
		this.apiService.JoinGameByCode(this.playerName, this.playerImg, this.joinCode).subscribe(this.joinResp())
	}

	joinPublicGame() {
		if (this.playerName == '') {
			this.showInvalidName()
			return
		}
		this.apiService.JoinPublicGame(this.playerName, this.playerImg).subscribe(this.joinResp())
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

	toggleJoinCodeInput() {
		if (this.showJoinCodeInput && this.joinCode && this.playerName) {
			this.joinGameByCode()
			return
		}
		this.showJoinCodeInput = !this.showJoinCodeInput
	}
}
