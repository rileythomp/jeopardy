import { Component, Input } from '@angular/core';

@Component({
    selector: 'app-lobby',
    templateUrl: './lobby.component.html',
    styleUrls: ['./lobby.component.less']
})
export class LobbyComponent {
    @Input() lobbyMessage: string;
    @Input() gameName: string;

    constructor() { }
}
