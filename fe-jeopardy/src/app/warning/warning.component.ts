import { Component, OnInit } from '@angular/core';

@Component({
	selector: 'app-warning',
	templateUrl: './warning.component.html',
	styleUrls: ['./warning.component.less']
})
export class WarningComponent implements OnInit {
	protected message: string;

	constructor() { }

	ngOnInit() {
		this.message = history.state.message;
	}
}
