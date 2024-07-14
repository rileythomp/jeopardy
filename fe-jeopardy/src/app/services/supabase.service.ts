import { Injectable } from '@angular/core';
import { SupabaseClient, createClient } from '@supabase/supabase-js';
import { environment } from 'src/environments/environment';

@Injectable({
	providedIn: 'root'
})
export class SupabaseService {
	private supabase: SupabaseClient<any, "public", any>

	constructor() {
		this.supabase = createClient(environment.supabaseUrl, environment.supabaseKey)
	}

	Storage() {
		return this.supabase.storage
	}

	Auth() {
		return this.supabase.auth
	}
}