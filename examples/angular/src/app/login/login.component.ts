import { Component } from '@angular/core';
import { environment } from '../../environments/environment';
import { Router } from '@angular/router';
import { register } from '@teamhanko/hanko-elements';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['../app.component.css']
})
export class LoginComponent {
  api = environment.hankoApi;
  error: Error | undefined;

  constructor(private router: Router) {
    register({ shadow: true }).catch((e) => this.error = e);
  }

  redirectToTodo() {
    this.router.navigate(['/todo']).catch((e) => this.error = e);
  }
}
