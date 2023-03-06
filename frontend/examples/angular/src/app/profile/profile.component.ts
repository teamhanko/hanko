import { Component } from '@angular/core';
import { environment } from '../../environments/environment';
import { Router } from '@angular/router';
import { register } from '@teamhanko/hanko-elements';
import { TodoService } from '../services/todo.service';

@Component({
  selector: 'app-profile',
  templateUrl: './profile.component.html',
  styleUrls: ['../app.component.css'],
})
export class ProfileComponent {
  api = environment.hankoApi;
  error: Error | undefined;

  constructor(private todoService: TodoService, private router: Router) {
    register({ shadow: true }).catch((e) => (this.error = e));
  }

  todos() {
    this.router.navigate(['/todo']).catch((e) => (this.error = e));
  }

  logout() {
    this.todoService
      .logout()
      .then(() => {
        this.router.navigate(['/']).catch((e) => (this.error = e));
        return;
      })
      .catch((e) => (this.error = e));
  }
}
