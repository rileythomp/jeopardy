import { Component, ElementRef, OnInit, ViewChild } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';
import { ModalService } from 'src/app/services/modal.service';

@Component({
	selector: 'app-password-reset',
	templateUrl: './password-reset.component.html',
	styleUrl: './password-reset.component.less'
})
export class PasswordResetComponent implements OnInit {
	protected password: string = ''
	protected confirmedPassword: string = ''
	protected showEye: boolean = true
	protected canReset: boolean = false
	@ViewChild('passwordInput') passwordInput: ElementRef
	@ViewChild('confirmedPasswordInput') confirmedPasswordInput: ElementRef

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
		if (!this.verifyPasswords()) {
			this.passwordsBorder('1px solid red')
			return
		}
		let error = await this.auth.UpdateUserPassword(this.password)
		if (error) {
			this.modal.displayMessage('Uh oh, there was an error restting your password. Please try again later.')
			return
		}
		error = await this.auth.SignOut()
		if (error) {
			return
		}
		location.replace('')
	}

	protected passwordsBorder(border: string) {
		this.passwordInput.nativeElement.style.border = border
		this.confirmedPasswordInput.nativeElement.style.border = border
	}

	private verifyPasswords(): boolean {
		return this.password.length > 7 &&
			this.password == this.confirmedPassword &&
			this.hasUppercase() && this.hasLowercase() &&
			this.hasNumber() && this.hasSpecialChar()
	}

	protected hasUppercase(): boolean {
		return /[A-Z]/.test(this.password)
	}

	protected hasLowercase(): boolean {
		return /[a-z]/.test(this.password)
	}

	protected hasNumber(): boolean {
		return /[0-9]/.test(this.password)
	}

	protected hasSpecialChar(): boolean {
		return /[^A-Za-z0-9]/.test(this.password)
	}

	protected showPassword(show: boolean) {
		this.passwordInput.nativeElement.type = show ? 'text' : 'password'
		this.confirmedPasswordInput.nativeElement.type = this.passwordInput.nativeElement.type
		this.showEye = !show
	}
}
