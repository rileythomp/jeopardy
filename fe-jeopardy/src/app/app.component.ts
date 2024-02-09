import { Component } from '@angular/core';
import { Router } from '@angular/router'

@Component({
	selector: 'app-root',
	templateUrl: './app.component.html',
	styleUrls: ['./app.component.less']
})
export class AppComponent {
	constructor(
		private router: Router,
	) { }

	ngOnInit() {
		if (window.innerHeight < 600 || window.innerWidth < 1140) {
			this.router.navigate(['/warning'], { state: { message: 'Your screen is to small to play this game. Please try on a larger screen.' } })
		}
	}
}
