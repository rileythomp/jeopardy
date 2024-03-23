import { Component, ElementRef, ViewChild } from '@angular/core';
import { AuthService } from 'src/app/services/auth.service';
import { ModalService } from 'src/app/services/modal.service';
import { StorageService } from 'src/app/services/storage.service';

@Component({
	selector: 'app-register',
	templateUrl: './register.component.html',
	styleUrl: './register.component.less'
})
export class RegisterComponent {
	protected uploadedImg: boolean = false;
	protected imgUpload: boolean = false;
	protected uploadedImgUrl: string = '';
	protected password: string = '';
	protected confirmedPassword = '';
	protected email: string = '';
	protected username: string = '';
	protected showEye: boolean = true
	protected registerDone: boolean = false
	@ViewChild('userImg') userImg: ElementRef
	@ViewChild('usernameInput') usernameInput: ElementRef
	@ViewChild('emailInput') emailInput: ElementRef
	@ViewChild('passwordInput') passwordInput: ElementRef
	@ViewChild('confirmedPasswordInput') confirmedPasswordInput: ElementRef

	constructor(
		private storage: StorageService,
		private auth: AuthService,
		protected modal: ModalService,
	) { }

	showImgUpload() {
		this.imgUpload = !this.imgUpload
	}

	async handleUserUpload(event: any) {
		let file = event.target.files[0]
		let { url, error } = await this.storage.UploadImg(file)
		if (error) {
			console.error(error)
			this.userImg.nativeElement.value = ''
			this.userImg.nativeElement.style.border = '1px solid red'
			setTimeout(() => {
				this.userImg.nativeElement.style.border = '1px solid grey'
			}, 2000)
			return
		}
		this.uploadedImg = true
		this.uploadedImgUrl = url
	}

	private validUsername(): boolean {
		return this.username.length > 0 && this.username.length < 50
	}

	private validEmail(): boolean {
		return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(this.email)
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

	protected usernameBorder(border: string) {
		this.usernameInput.nativeElement.style.border = border
	}

	protected emailBorder(border: string) {
		this.emailInput.nativeElement.style.border = border
	}

	protected passwordsBorder(border: string) {
		this.passwordInput.nativeElement.style.border = border
		this.confirmedPasswordInput.nativeElement.style.border = border
	}

	async createAccount() {
		let hasError = false
		if (!this.validUsername()) {
			this.usernameBorder('1px solid red')
			hasError = true
		}
		if (!this.validEmail()) {
			this.emailBorder('1px solid red')
			hasError = true
		}
		if (!this.verifyPasswords()) {
			this.passwordsBorder('1px solid red')
			hasError = true
		}
		if (hasError) {
			return
		}
		let error = await this.auth.SignUp(this.email, this.password, this.username, this.uploadedImgUrl)
		if (error) {
			console.error(error)
			this.modal.displayMessage('Uh oh, we\'re not able to create new accounts right now. Please try again later.')
			return
		}
		this.usernameInput.nativeElement.disabled = true
		this.emailInput.nativeElement.disabled = true
		this.showPassword(false)
		this.passwordInput.nativeElement.disabled = true
		this.confirmedPasswordInput.nativeElement.disabled = true
		this.registerDone = true
	}

	protected showPassword(show: boolean) {
		this.passwordInput.nativeElement.type = show ? 'text' : 'password'
		this.confirmedPasswordInput.nativeElement.type = this.passwordInput.nativeElement.type
		this.showEye = !show
	}
}
