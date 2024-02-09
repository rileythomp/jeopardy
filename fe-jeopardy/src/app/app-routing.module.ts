import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { GameComponent } from './game/game.component';
import { JoinComponent } from './join/join.component';
import { LinkJoinComponent } from './link-join/link-join.component';
import { WarningComponent } from './warning/warning.component';

const routes: Routes = [
	{ path: '', component: JoinComponent },
	{ path: 'join', component: JoinComponent },
	{ path: 'join/:gameCode', component: LinkJoinComponent },
	{ path: 'game/:gameCode', component: GameComponent },
	{ path: 'warning', component: WarningComponent }

];

@NgModule({
	imports: [RouterModule.forRoot(routes)],
	exports: [RouterModule]
})
export class AppRoutingModule { }
