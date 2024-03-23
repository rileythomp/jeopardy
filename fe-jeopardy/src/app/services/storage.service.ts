import { Injectable } from '@angular/core';
import { environment } from 'src/environments/environment';
import { SupabaseService } from './supabase.service';

@Injectable({
	providedIn: 'root'
})
export class StorageService {

	constructor(private supabase: SupabaseService) { }

	public async UploadImg(file: File): Promise<{ error: Error | null, url: string }> {
		let { data, error } = await this.supabase.Storage().from('jeopardy_user_imgs').upload(file.name, file)
		if (error) {
			console.error(error)
			return { error: error, url: '' }
		}
		let url = environment.supabaseUrl + '/storage/v1/object/public/jeopardy_user_imgs/' + data?.path
		return { error: null, url: url }
	}
}
