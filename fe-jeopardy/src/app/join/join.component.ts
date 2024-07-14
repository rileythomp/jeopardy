import { Component, OnInit } from '@angular/core'
import { Router } from '@angular/router'
import { Observer } from 'rxjs'
import { ServerUnavailableMsg, User } from '../model/model'
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
	protected user: User
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
			this.user = user
		})
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
		if (this.user.name == '') {
			this.showInvalidName()
			return
		}
		this.api.CreatePrivateGame(
			this.user.name, this.user.imgUrl, this.user.email,
			bots, this.twoRoundChecked, this.penaltyChecked,
			30, 30, 15, 30, [], []
		).subscribe(this.joinResp())
	}

	joinGameByCode() {
		if (this.user.name == '') {
			this.showInvalidName()
			return
		}
		this.api.JoinGameByCode(this.user.name, this.user.imgUrl, this.user.email, this.joinCode).subscribe(this.joinResp())
	}

	joinPublicGame() {
		if (this.user.name == '') {
			this.showInvalidName()
			return
		}
		this.api.JoinPublicGame(this.user.name, this.user.imgUrl, this.user.email).subscribe(this.joinResp())
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
		if (this.showJoinCodeInput && this.joinCode && this.user.name) {
			this.joinGameByCode()
			return
		}
		this.showJoinCodeInput = !this.showJoinCodeInput
	}
}
