import { AfterViewInit, Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { environment } from 'src/environments/environment';
import { AuthService } from './services/auth.service';
import { ModalService } from './services/modal.service';

@Component({
	selector: 'app-root',
	templateUrl: './app.component.html',
	styleUrls: ['./app.component.less']
})
export class AppComponent implements OnInit, AfterViewInit {
	protected showLoginOptions: boolean = false
	protected showLogoutOptions: boolean = false
	protected userAuthenticated: boolean = false
	protected userImg: string = ''

	constructor(
		private router: Router,
		protected modal: ModalService,
		protected auth: AuthService,
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

Please report any issues at https://github.com/rileythomp/jeopardy/issues/new
`
		console.log(jeopardy)

		this.auth.user.subscribe(user => {
			this.userAuthenticated = user.authenticated
			this.userImg = user.imgUrl
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
}
