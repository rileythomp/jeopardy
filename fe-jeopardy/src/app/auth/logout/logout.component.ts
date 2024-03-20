import { Component } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';

@Component({
	selector: 'app-logout',
	templateUrl: './logout.component.html',
	styleUrls: ['./logout.component.less']
})
export class LogoutComponent {

	constructor(private auth: AuthService) { }

	async googleLogout() {
		await this.auth.SignOut();
		location.reload();
	}
}
