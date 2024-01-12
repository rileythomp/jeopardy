import { Injectable } from '@angular/core';
import { environment } from 'src/environments/environment';

@Injectable({
	providedIn: 'root'
})
export class WebsocketService {
	private ws: WebSocket;

	constructor() { }

	connect(path: string): void {
		this.ws = new WebSocket(`${environment.websocketProtocol}://${environment.apiServerUrl}/jeopardy/${path}`);
	}

	onopen(callback: () => void): void {
		this.ws.onopen = callback;
	}

	send(data: any): void {
		this.ws.send(JSON.stringify(data));
	}

	onmessage(callback: (event: { data: string }) => void): void {
		this.ws.onmessage = callback;
	}
}
