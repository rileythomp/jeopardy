import { Component, EventEmitter, Output } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';

@Component({
	selector: 'app-sign-out',
	templateUrl: './sign-out.component.html',
	styleUrls: ['./sign-out.component.less']
})
export class SignOutComponent {
	@Output() signOutError = new EventEmitter<boolean>();

	constructor(private auth: AuthService) { }

	async googleSignOut() {
		if (await this.auth.SignOut()) {
			this.signOutError.emit(true);
		} else {
			location.reload();
		}
	}
}
