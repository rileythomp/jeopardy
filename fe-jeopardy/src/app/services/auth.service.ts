import { Injectable } from '@angular/core';
import { SupabaseClient, createClient } from '@supabase/supabase-js';
import { User } from '../model/model';

@Injectable({
	providedIn: 'root'
})
export class AuthService {
	private supabase: SupabaseClient<any, "public", any>
	private user: User

	constructor() {
		this.supabase = createClient('https://xdlhyjzjygansfeoguvs.supabase.co', 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InhkbGh5anpqeWdhbnNmZW9ndXZzIiwicm9sZSI6ImFub24iLCJpYXQiOjE3MDY5OTUwMjksImV4cCI6MjAyMjU3MTAyOX0.ystMHS-Tic8W3rHqXTwW1F90WvxfVHpLJ5bkimn81PM');
	}

	public async GetUser() {
		console.log('getting user')
		let { data, error } = await this.supabase.auth.getUser();
		if (error) {
			console.log(error)
			console.log('user is not signed in')
			this.user = <User>{}
		} else {
			console.log(data)
			console.log('user is signed in')
			let user: User = {
				imgUrl: data.user?.user_metadata['avatar_url'],
				authenticated: true,
				name: data.user?.user_metadata['full_name']
			}
			this.user = user
		}
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
		} else {
			console.log(data)
			console.log('signed in successfully')
		}
	}

	public async SignOut() {
		console.log('signing out')
		let { error } = await this.supabase.auth.signOut();
		if (error) {
			console.log(error)
			console.log('there was an error signing out')
		} else {
			console.log('signed out successfully')
		}
	}

	public Authenticated(): boolean {
		return this.user?.authenticated
	}

	public ImgUrl(): string {
		return this.user?.imgUrl
	}

	public Name(): string {
		return this.user?.name
	}
}
