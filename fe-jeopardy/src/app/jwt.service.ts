import { Injectable } from '@angular/core';
import { BehaviorSubject, Observable } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class JwtService {
  private jwtSubject: BehaviorSubject<string>;
  public jwt$: Observable<string>;

  constructor() {
    const storedJwt:string = localStorage.getItem('jwt') ?? '';
    this.jwtSubject = new BehaviorSubject<string>(storedJwt);
    this.jwt$ = this.jwtSubject.asObservable();
  }

  public setJwt(jwt: string): void {
    localStorage.setItem('jwt', jwt);
    this.jwtSubject.next(jwt);
  }

  public clearJwt(): void {
    localStorage.removeItem('jwt');
    this.jwtSubject.next('');
  }

  public getJwt(): string {
    return this.jwtSubject.value;
  }

  public hasJwt(): boolean {
    return !!this.jwtSubject.value;
  }
}
