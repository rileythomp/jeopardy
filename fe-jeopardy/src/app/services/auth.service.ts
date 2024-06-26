import { Injectable } from '@angular/core';
import { Provider } from '@supabase/supabase-js';
import { BehaviorSubject, Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { User } from '../model/model';
import { SupabaseService } from './supabase.service';
import { formattedDate } from './utils';

@Injectable({
	providedIn: 'root'
})
export class AuthService {
	private userSubject: BehaviorSubject<User>
	public user: Observable<User>

	constructor(private supabase: SupabaseService) {
		this.userSubject = new BehaviorSubject<User>({
			email: '',
			imgUrl: '',
			authenticated: false,
			name: '',
			dateJoined: '',
			public: false,
		});
		this.user = this.userSubject.asObservable();
	}

	public async UpdateUserName(name: string): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().updateUser({
			data: {
				display_name: name,
			}
		})
		if (error) {
			console.error(error)
			return error
		}
		return null
	}

	public async UpdateProfileVisibility(profilePublic: boolean): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().updateUser({
			data: {
				profile_public: profilePublic,
			}
		})
		if (error) {
			console.error(error)
			return error
		}
		return null
	}

	public async UpdateUserImg(url: string): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().updateUser({
			data: {
				user_img_url: url,
			}
		})
		if (error) {
			console.error(error)
			return error
		}
		return null
	}

	public async UpdateUserPassword(password: string): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().updateUser({
			password: password,
		})
		if (error) {
			console.error(error)
			return error
		}
		return null
	}

	public async SendPasswordResetEmail(email: string): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().resetPasswordForEmail(email, {
			redirectTo: environment.passwordRedirectUrl,
		})
		if (error) {
			console.error(error)
			return error
		}
		return null
	}

	public async GetUser(): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().getUser();
		if (error) {
			return error
		}
		let user: User = {
			email: data.user?.email ?? '',
			imgUrl: data.user?.user_metadata['user_img_url'] ?? data.user?.user_metadata['avatar_url'],
			authenticated: true,
			name: data.user?.user_metadata['display_name'] ?? data.user?.user_metadata['full_name'],
			dateJoined: formattedDate(data.user?.created_at ?? ''),
			public: data.user?.user_metadata['profile_public'] ?? true,
		}
		this.userSubject.next(user)
		return null
	}

	public async SignUp(email: string, password: string, username: string, imgUrl: string): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().signUp({
			email: email,
			password: password,
			options: {
				emailRedirectTo: environment.redirectUrl,
				data: {
					full_name: username,
					avatar_url: imgUrl,
				}
			}
		})
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

	public async SignInWithPassword(email: string, password: string): Promise<Error | null> {
		let { data, error } = await this.supabase.Auth().signInWithPassword({
			email: email,
			password: password,
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
}
