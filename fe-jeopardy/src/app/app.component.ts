import { Component, OnInit, AfterViewInit } from '@angular/core';
import { Router } from '@angular/router'
import { environment } from 'src/environments/environment';
import { ModalService } from './services/modal.service';

@Component({
	selector: 'app-root',
	templateUrl: './app.component.html',
	styleUrls: ['./app.component.less']
})
export class AppComponent implements OnInit, AfterViewInit {
	constructor(
		private router: Router,
		protected modal: ModalService,
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
