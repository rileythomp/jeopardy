import { Injectable } from '@angular/core';
import { environment } from 'src/environments/environment';

@Injectable({
	providedIn: 'root'
})
export class ChatService {
	private ws: WebSocket;

	constructor() { }

	Connect(): void {
		this.ws = new WebSocket(`${environment.websocketProtocol}://${environment.apiServerUrl}/jeopardy/chat`);
	}

	OnOpen(callback: () => void): void {
		this.ws.onopen = callback;
	}

	Send(data: any): void {
		this.ws.send(JSON.stringify(data));
	}

	OnMessage(callback: (event: { data: string }) => void): void {
		this.ws.onmessage = callback;
	}
}
