import { Component, EventEmitter, Output } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';

@Component({
	selector: 'app-logout',
	templateUrl: './logout.component.html',
	styleUrls: ['./logout.component.less']
})
export class LogoutComponent {
	@Output() signOutError = new EventEmitter<boolean>();

	constructor(private auth: AuthService) { }

	async googleLogout() {
		if (await this.auth.SignOut()) {
			this.signOutError.emit(true);
		} else {
			location.reload();
		}
	}
}
