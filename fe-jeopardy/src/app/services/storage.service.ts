import { Injectable } from '@angular/core';
import { environment } from 'src/environments/environment';
import { v4 as uuidv4 } from 'uuid';
import { SupabaseService } from './supabase.service';

const userImgs: string = 'jeopardy_user_imgs'

@Injectable({
	providedIn: 'root'
})
export class StorageService {

	constructor(private supabase: SupabaseService) { }

	public async UploadImg(folder: string, file: File): Promise<{ error: Error | null, url: string }> {
		let { data, error } = await this.supabase.Storage().from(userImgs).upload(`${folder}/${uuidv4()}-${file.name}`, file)
		if (error) {
			console.error(error)
			return { error: error, url: '' }
		}
		let url = `${environment.supabaseUrl}/storage/v1/object/public/${userImgs}/${data?.path}`
		return { error: null, url: url }
	}
}
