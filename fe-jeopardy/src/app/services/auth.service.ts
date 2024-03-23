import { Injectable } from '@angular/core';
import { Provider, SignInWithPasswordCredentials, SignUpWithPasswordCredentials } from '@supabase/supabase-js';
import { Observable, Subject } from 'rxjs';
import { environment } from 'src/environments/environment';
import { User } from '../model/model';
import { SupabaseService } from './supabase.service';

@Injectable({
	providedIn: 'root'
})
export class AuthService {
	private userSubject: Subject<User>
	public user: Observable<User>

	constructor(private supabase: SupabaseService) {
		this.userSubject = new Subject<User>();
		this.user = this.userSubject.asObservable();
	}

	public async GetUser() {
		let { data, error } = await this.supabase.Auth().getUser();
		if (error) {
			return
		}
		let user: User = {
			imgUrl: data.user?.user_metadata['avatar_url'],
			authenticated: true,
			name: data.user?.user_metadata['full_name']
		}
		this.userSubject.next(user)
	}

	public async SignUp(credentials: SignUpWithPasswordCredentials): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().signUp(credentials)
		if (error) {
			console.error(error)
			return error
		}
		return null
	}

	public async SignIn(provider: string): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().signInWithOAuth({
			provider: provider as Provider,
			options: {
				redirectTo: environment.redirectUrl,
			}
		})
		if (error) {
			console.error(error)
			return error
		}
		return null
	}

	public async SignOut(): Promise<Error | null> {
		let { error } = await this.supabase.Auth().signOut();
		if (error) {
			console.error(error)
			return error
		}
		return null
	}

	public async SignInWithPassword(credentials: SignInWithPasswordCredentials): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().signInWithPassword(credentials)
		if (error) {
			console.error(error)
			return error
		}
		return null
	}
}
