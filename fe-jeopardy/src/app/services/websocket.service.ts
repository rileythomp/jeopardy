import { Injectable } from '@angular/core';
import { environment } from 'src/environments/environment';

@Injectable({
	providedIn: 'root'
})
export class WebsocketService {
	private ws: WebSocket;

	constructor() { }

	Connect(path: string): void {
		this.ws = new WebSocket(`${environment.websocketProtocol}://${environment.apiServerUrl}/jeopardy/${path}`);
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
