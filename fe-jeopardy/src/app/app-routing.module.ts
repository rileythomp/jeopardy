import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { PasswordResetComponent } from './auth/password-reset/password-reset.component';
import { GameComponent } from './game/game.component';
import { JoinComponent } from './join/join.component';
import { LeaderboardsComponent } from './leaderboards/leaderboards.component';
import { LinkJoinComponent } from './link-join/link-join.component';
import { ProfileComponent } from './profile/profile.component';
import { WarningComponent } from './warning/warning.component';

const routes: Routes = [
	{ path: '', component: JoinComponent },
	{ path: 'join', component: JoinComponent },
	{ path: 'join/:joinCode', component: LinkJoinComponent },
	{ path: 'game/:joinCode', component: GameComponent },
	{ path: 'warning', component: WarningComponent },
	{ path: 'profile', component: ProfileComponent },
	{ path: 'profile/:name', component: ProfileComponent },
	{ path: 'password-reset', component: PasswordResetComponent },
	{ path: 'leaderboards', component: LeaderboardsComponent },
];

@NgModule({
	imports: [RouterModule.forRoot(routes)],
	exports: [RouterModule]
})
export class AppRoutingModule { }
