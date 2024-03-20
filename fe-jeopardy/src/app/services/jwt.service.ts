import { Injectable } from '@angular/core';

const jeopardyJWT = 'jeopardyJWT';

@Injectable({
	providedIn: 'root'
})
export class JwtService {
	constructor() { }

	SetJWT(jwt: string): void {
		localStorage.setItem(jeopardyJWT, jwt);
	}

	GetJWT(): string {
		return localStorage.getItem(jeopardyJWT) ?? '';
	}
}
