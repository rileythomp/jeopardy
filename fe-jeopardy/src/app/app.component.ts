import { animate, state, style, transition, trigger } from '@angular/animations';
import { AfterViewInit, Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { environment } from 'src/environments/environment';
import { AuthService } from './services/auth.service';
import { ModalService } from './services/modal.service';

@Component({
	selector: 'app-root',
	templateUrl: './app.component.html',
	styleUrls: ['./app.component.less'],
	animations: [
		trigger('slideDownUp', [
			state('void', style({ transform: 'scaleY(0)' })),
			state('*', style({ transform: 'scaleY(1)' })),
			transition('* <=> void', animate('0.3s')),
		]),
	],
})
export class AppComponent implements OnInit, AfterViewInit {
	protected showAuthOptions: boolean = false
	protected userAuthenticated: boolean = false
	protected playerImg: string = ''
	protected showRegistration: boolean = false

	constructor(
		private router: Router,
		protected modal: ModalService,
		private auth: AuthService,
	) { }

	ngOnInit() {
		if (environment.production && (window.innerHeight < 600 || window.innerWidth < 1140)) {
			this.router.navigate(['/warning'], { state: { message: 'Your screen is to small to play this game. Please try on a larger screen.' } })
		}
		let jeopardy =
			`   ___                                _       
  |_  |                              | |      
    | | ___  ___  _ __   __ _ _ __ __| |_   _ 
	| |/ _ \\/ _ \\|  _ \\ / _  |  __/ _  | | | |
/\\__/ /  __/ (_) | |_) | (_| | | | (_| | |_| |
\\____/ \\___|\\___/|  __/ \\____|_|  \\____|\\___ |
			     | |                     __/ |
			     |_|                    |___/ 		

Please report any issues to https://docs.google.com/forms/d/e/1FAIpQLSdzHFumIhdsgNksr8lDUO3hHhVwaIqeO9asIhBWsroNfYZW4Q/viewform
`
		console.log(jeopardy)

		this.auth.user.subscribe(user => {
			this.userAuthenticated = user.authenticated
			this.playerImg = user.imgUrl
		})

		this.auth.GetUser()
	}

	ngAfterViewInit() {
		if (environment.production && (window.innerHeight < 600 || window.innerWidth < 1140)) {
			return
		}
		let showJeopardyInfo = localStorage.getItem('showJeopardyInfo')
		if (showJeopardyInfo === null) {
			localStorage.setItem('showJeopardyInfo', 'shown')
			this.modal.displayInstructions()
		}
	}

	async signIn(provider: string) {
		if (await this.auth.SignIn(provider)) {
			this.handleAuthError('Uh oh, there was an unexpected error signing in. Please try again.')
		}
	}

	async signOut() {
		if (await this.auth.SignOut()) {
			this.handleAuthError('Uh oh, there was an unexpected error signing out. Please try again.')
			return
		}
		location.replace('');
	}

	startRegistration() {
		this.showAuthOptions = false
		this.modal.displayRegister()
	}

	handleAuthError(msg: string) {
		this.showAuthOptions = false
		this.modal.displayMessage(msg)
	}
}
