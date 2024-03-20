import { Injectable } from '@angular/core';
import { SupabaseClient, createClient } from '@supabase/supabase-js';
import { BehaviorSubject, Observable } from 'rxjs';
import { User } from '../model/model';

@Injectable({
	providedIn: 'root'
})
export class AuthService {
	private supabase: SupabaseClient<any, "public", any>
	private userSubject: BehaviorSubject<User | null>
	public user: Observable<User | null>


	constructor() {
		this.supabase = createClient('https://xdlhyjzjygansfeoguvs.supabase.co', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InhkbGh5anpqeWdhbnNmZW9ndXZzIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MDY5OTUwMjksImV4cCI6MjAyMjU3MTAyOX0.ystMHS-Tic8W3rHqXTwW1F90WvxfVHpLJ5bkimn81PM');
		this.userSubject = new BehaviorSubject<User | null>(null);
		this.user = this.userSubject.asObservable();
	}

	public async GetUser() {
		let { data, error } = await this.supabase.auth.getUser();
		if (error) {
			console.log(error)
			console.log('user is not signed in')
			return
		}
		console.log(data)
		console.log('user is signed in')
		let user: User = {
			imgUrl: data.user?.user_metadata['avatar_url']
		}
		this.userSubject.next(user)
	}

	public async SignIn() {
		console.log('signing in')
		let { data, error } = await this.supabase.auth.signInWithOAuth({
			provider: 'google',
			options: {
				redirectTo: 'http://localhost:4200/join',
			}
		})
		if (error) {
			console.log(error)
			console.log('there was an error signing in')
			return
		}
		console.log(data)
		console.log('signed in successfully')
		let user: User = {
			imgUrl: ""
		}
		this.userSubject.next(user)
	}

	public async SignOut() {
		console.log('sign out')
		let { error } = await this.supabase.auth.signOut();
		if (error) {
			console.log(error)
			console.log('there was an error signing out')
			return
		}
		this.userSubject.next(null)
	}
}
