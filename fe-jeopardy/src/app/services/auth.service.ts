import { Injectable } from '@angular/core';
import { SupabaseClient, createClient } from '@supabase/supabase-js';
import { Observable, Subject } from 'rxjs';
import { User } from '../model/model';

@Injectable({
	providedIn: 'root'
})
export class AuthService {
	private supabase: SupabaseClient<any, "public", any>
	private userSubject: Subject<User>
	public user: Observable<User>

	constructor() {
		this.supabase = createClient('https://xdlhyjzjygansfeoguvs.supabase.co', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InhkbGh5anpqeWdhbnNmZW9ndXZzIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MDY5OTUwMjksImV4cCI6MjAyMjU3MTAyOX0.ystMHS-Tic8W3rHqXTwW1F90WvxfVHpLJ5bkimn81PM');
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

	public async SignIn(): Promise<Error | null> {
		let { data, error } = await this.supabase.auth.signInWithOAuth({
			provider: 'google',
			options: {
				redirectTo: 'http://localhost:4200/join',
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
