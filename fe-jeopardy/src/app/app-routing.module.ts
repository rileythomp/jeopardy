import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { GameComponent } from './game/game.component';
import { JoinComponent } from './join/join.component';
import { LinkJoinComponent } from './link-join/link-join.component';

const routes: Routes = [
	{ path: '', component: JoinComponent },
	{ path: 'join', component: JoinComponent },
	{ path: 'game', component: GameComponent },
	{ path: ':gameCode', component: LinkJoinComponent }

];

@NgModule({
	imports: [RouterModule.forRoot(routes)],
	exports: [RouterModule]
})
export class AppRoutingModule { }
