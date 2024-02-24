import { Component, OnInit, AfterViewInit, ViewChild } from '@angular/core';
import { Router } from '@angular/router'
import { ModalComponent } from './modal/modal.component';

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

	ngOnInit() {
		if (window.innerHeight < 600 || window.innerWidth < 1140) {
			this.router.navigate(['/warning'], { state: { message: 'Your screen is to small to play this game. Please try on a larger screen.' } })
		}
	}

	ngAfterViewInit() {
		if (window.innerHeight < 600 || window.innerWidth < 1140) {
			return
		}
		let showJeopardyInfo = localStorage.getItem('showJeopardyInfo')
		if (showJeopardyInfo === null) {
			localStorage.setItem('showJeopardyInfo', 'shown')
			this.modalComponent.showJeopardyInfo()
		}
	}
}
