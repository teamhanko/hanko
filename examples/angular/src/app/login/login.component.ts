import { Component } from '@angular/core';
import { environment } from '../../environments/environment';
import { Router } from '@angular/router';
import { register } from '@teamhanko/hanko-elements/hanko-auth';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css'],
})
export class LoginComponent {
  api = environment.hankoApi;

  constructor(private router: Router) {
    register({ shadow: true }).catch((e) => console.error(e));
  }

  redirectToTodo() {
    this.router.navigate(['/todo']);
  }
}
