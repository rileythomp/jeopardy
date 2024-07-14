import { Component, ElementRef, Input, ViewChild } from '@angular/core';

@Component({
	selector: 'app-new-password',
	templateUrl: './new-password.component.html',
	styleUrl: './new-password.component.less'
})
export class NewPasswordComponent {
	password: string = ''
	confirmedPassword: string = ''
	protected showEye: boolean = true
	@Input() inputFlexDirection: string
	@ViewChild('passwordInput') passwordInput: ElementRef
	@ViewChild('confirmedPasswordInput') confirmedPasswordInput: ElementRef

	public isValid(): boolean {
		return this.password.length > 7 &&
			this.password == this.confirmedPassword &&
			this.hasUppercase() && this.hasLowercase() &&
			this.hasNumber()
	}

	public setBorder(border: string) {
		this.passwordInput.nativeElement.style.border = border
		this.confirmedPasswordInput.nativeElement.style.border = border
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

	protected showPassword(show: boolean) {
		this.passwordInput.nativeElement.type = show ? 'text' : 'password'
		this.confirmedPasswordInput.nativeElement.type = this.passwordInput.nativeElement.type
		this.showEye = !show
	}

	public disableInput(): void {
		this.showPassword(false)
		this.passwordInput.nativeElement.disabled = true
		this.confirmedPasswordInput.nativeElement.disabled = true
	}
}
