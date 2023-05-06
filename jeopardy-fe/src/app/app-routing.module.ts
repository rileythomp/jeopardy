import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { GameComponent } from './game/game.component';
import { JoinComponent } from './join/join.component';
import { LobbyComponent } from './lobby/lobby.component';

const routes: Routes = [
	{ path: '', component: JoinComponent },
	{ path: 'lobby', component: LobbyComponent },
	{ path: 'game', component: GameComponent },

];

@NgModule({
	imports: [RouterModule.forRoot(routes)],
	exports: [RouterModule]
})
export class AppRoutingModule { }
