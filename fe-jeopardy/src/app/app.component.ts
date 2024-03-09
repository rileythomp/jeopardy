import { Component, OnInit, AfterViewInit, ViewChild } from '@angular/core';
import { Router } from '@angular/router'
import { ModalComponent } from './modal/modal.component';
import { AnalyticsComponent } from './analytics/analytics.component';

@Component({
	selector: 'app-root',
	templateUrl: './app.component.html',
	styleUrls: ['./app.component.less']
})
export class AppComponent implements OnInit, AfterViewInit {
	constructor(
		private router: Router,
	) { }

	@ViewChild(ModalComponent) modalComponent: ModalComponent
	@ViewChild(AnalyticsComponent) analytics: AnalyticsComponent

	ngOnInit() {
		// if (window.innerHeight < 600 || window.innerWidth < 1140) {
		// 	this.router.navigate(['/warning'], { state: { message: 'Your screen is to small to play this game. Please try on a larger screen.' } })
		// }
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
		if (window.innerHeight < 600 || window.innerWidth < 1140) {
			return
		}
		let showJeopardyInfo = localStorage.getItem('showJeopardyInfo')
		if (showJeopardyInfo === null) {
			localStorage.setItem('showJeopardyInfo', 'shown')
			this.modalComponent.showJeopardyInfo(true, true)
		}
	}

	showAnalytics() {
		this.modalComponent.showJeopardyInfo(false, false)
		this.analytics.toggleAnalytics(true)
	}

	showJeopardyInfo() {
		this.analytics.toggleAnalytics(false)
		this.modalComponent.showJeopardyInfo(false, true)
	}
}
