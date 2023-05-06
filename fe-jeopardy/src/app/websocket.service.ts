import { Injectable } from '@angular/core';

@Injectable({
	providedIn: 'root'
})
export class WebsocketService {
	private ws: WebSocket;

	constructor() { }

	connect(url: string): void {
		this.ws = new WebSocket(url);
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
