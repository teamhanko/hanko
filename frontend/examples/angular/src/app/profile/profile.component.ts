import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { TodoService } from '../services/todo.service';
import { HankoService } from "../services/hanko.services";

@Component({
  selector: 'app-profile',
  templateUrl: './profile.component.html',
  styleUrls: ['../app.component.css'],
})
export class ProfileComponent implements OnInit {
  error?: Error;

  constructor(private hankoService: HankoService, private todoService: TodoService, private router: Router) {}

  async ngOnInit() {
    const { is_valid} = await this.hankoService.client.validateSession();
    if (!is_valid) {
      this.redirectToLogin();
    }
  }

  logout() {
    this.hankoService.client.logout().catch((e) => (this.error = e));
  }

  redirectToLogin() {
    this.router.navigate(['/']).catch((e) => (this.error = e));
  }

  redirectToTodos() {
    this.router.navigate(['/todo']).catch((e) => (this.error = e));
  }
}
