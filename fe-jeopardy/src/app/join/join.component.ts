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
	protected playerImg: string = ''
	protected playerEmail: string = ''
	protected userAuthenticated: boolean = false
	protected joinCode: string = ''
	protected showJoinCodeInput: boolean = false
	protected oneRoundChecked: boolean = true
	protected twoRoundChecked: boolean = false
	protected penaltyChecked: boolean = true

	constructor(
		private router: Router,
		private jwt: JwtService,
		private api: ApiService,
		protected modal: ModalService,
		private auth: AuthService,
	) { }

	ngOnInit() {
		this.auth.user.subscribe(user => {
			this.userAuthenticated = user.authenticated
			this.playerName = user.name
			this.playerImg = user.imgUrl
			this.playerEmail = user.email
		})

		this.auth.GetUser()
	}

	private joinResp(): Partial<Observer<any>> {
		return {
			next: (resp: any) => {
				this.jwt.SetJWT(resp.token)
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
		this.api.CreatePrivateGame(
			this.playerName, this.playerImg, this.playerEmail,
			bots, this.twoRoundChecked, this.penaltyChecked,
			30, 30, 15, 30, [], []
		).subscribe(this.joinResp())
	}

	joinGameByCode() {
		if (this.playerName == '') {
			this.showInvalidName()
			return
		}
		this.api.JoinGameByCode(this.playerName, this.playerImg, this.playerEmail, this.joinCode).subscribe(this.joinResp())
	}

	joinPublicGame() {
		if (this.playerName == '') {
			this.showInvalidName()
			return
		}
		this.api.JoinPublicGame(this.playerName, this.playerImg, this.playerEmail).subscribe(this.joinResp())
	}

	rejoin() {
		this.api.GetPlayerGame().subscribe({
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
