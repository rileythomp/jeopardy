import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class WebsocketService {
  ws: WebSocket;

  constructor() {}

  connect(url: string): void {
    this.ws = new WebSocket(url);
  }
}
