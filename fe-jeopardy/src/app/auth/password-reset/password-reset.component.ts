import { Component, OnInit, ViewChild } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';
import { ModalService } from 'src/app/services/modal.service';
import { NewPasswordComponent } from '../new-password/new-password.component';

@Component({
	selector: 'app-password-reset',
	templateUrl: './password-reset.component.html',
	styleUrl: './password-reset.component.less'
})
export class PasswordResetComponent implements OnInit {
	protected canReset: boolean = false
	@ViewChild(NewPasswordComponent) passwords: NewPasswordComponent

	constructor(
		private auth: AuthService,
		private modal: ModalService,
	) { }

	async ngOnInit() {
		if (await this.auth.GetUser()) {
			location.replace('')
			return
		}
		this.canReset = true
	}

	async updatePassword() {
		if (!this.passwords.isValid()) {
			this.passwords.setBorder('1px solid red')
			return
		}
		if (await this.auth.UpdateUserPassword(this.passwords.password)) {
			this.modal.displayMessage('Uh oh, there was an error restting your password. Please try again later.')
			return
		}
		if (await this.auth.SignOut()) {
			return
		}
		location.replace('')
	}
}
