import { Injectable } from '@angular/core';
import { Provider, SupabaseClient, createClient } from '@supabase/supabase-js';
import { Observable, Subject } from 'rxjs';
import { environment } from 'src/environments/environment';
import { User } from '../model/model';

@Injectable({
	providedIn: 'root'
})
export class AuthService {
	private supabase: SupabaseClient<any, "public", any>
	private userSubject: Subject<User>
	public user: Observable<User>

	constructor() {
		this.supabase = createClient(environment.supabaseUrl, environment.supabaseKey);
		this.userSubject = new Subject<User>();
		this.user = this.userSubject.asObservable();
	}

	public async GetUser() {
		let { data, error } = await this.supabase.auth.getUser();
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

	public async SignIn(provider: string): Promise<Error | null> {
		let { data, error } = await this.supabase.auth.signInWithOAuth({
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
		let { error } = await this.supabase.auth.signOut();
		if (error) {
			console.error(error)
			return error
		}
		return null
	}
}
