import { Component, Input } from '@angular/core';

@Component({
    selector: 'app-pre-game',
    templateUrl: './pre-game.component.html',
    styleUrls: ['./pre-game.component.less']
})
export class PreGameComponent {
    @Input() preGameMessage: string;
    @Input() gameName: string;

    constructor() { }
}
