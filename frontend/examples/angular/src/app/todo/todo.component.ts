import { Component, CUSTOM_ELEMENTS_SCHEMA, OnInit, ViewChild } from "@angular/core";
import { Router } from '@angular/router';
import { Todos, TodoService } from '../services/todo.service';
import { HankoService } from "../services/hanko.services";
import { SessionExpiredModalComponent } from "../modal/session-expired-modal.component";

@Component({
    selector: 'app-todo',
    templateUrl: './todo.component.html',
    styleUrls: ['../app.component.css', './todo.component.css'],
    standalone: true,
    imports: [SessionExpiredModalComponent],
    schemas: [CUSTOM_ELEMENTS_SCHEMA]
})
export class TodoComponent implements OnInit {
  todos: Todos = [];
  error?: Error;
  description = '';

  constructor(private hankoService: HankoService, private todoService: TodoService, private router: Router) {}

  @ViewChild(SessionExpiredModalComponent)
  private sessionExpiredModalComponent!: SessionExpiredModalComponent;

  async ngOnInit() {
    const { is_valid} = await this.hankoService.client.validateSession();
    if (is_valid) {
        this.listTodos();
      } else {
        this.redirectToLogin();
      }
  }

  changeDescription(event: any) {
    this.description = event.target.value;
  }

  changeCheckbox(event: any) {
    const { currentTarget } = event;
    this.patchTodo(currentTarget.value, currentTarget.checked);
  }

  addTodo(event: any) {
    event.preventDefault();
    const entry = { description: this.description, checked: false };

    this.todoService
      .addTodo(entry)
      .then((res) => {
        if (res.status === 401) {
          this.sessionExpiredModalComponent.show();
          return;
        }

        this.description = '';
        this.listTodos();

        return;
      })
      .catch((e) => (this.error = e));
  }

  patchTodo(id: string, checked: boolean) {
    this.todoService
      .patchTodo(id, checked)
      .then((res) => {
        if (res.status === 401) {
          this.sessionExpiredModalComponent.show();
          return;
        }

        this.listTodos();

        return;
      })
      .catch((e) => (this.error = e));
  }

  listTodos() {
    this.todoService
      .listTodos()
      .then((res) => {
        if (res.status === 401) {
          this.sessionExpiredModalComponent.show();
          return;
        }

        return res.json();
      })
      .then((todo) => {
        if (todo) {
          this.todos = todo;
        }
      })
      .catch((e) => (this.error = e));
  }

  deleteTodo(id: string) {
    this.todoService
      .deleteTodo(id)
      .then((res) => {
        if (res.status === 401) {
          this.sessionExpiredModalComponent.show();
          return;
        }

        this.listTodos();

        return;
      })
      .catch((e) => (this.error = e));
  }

  logout() {
    this.hankoService.client.logout().catch((e) => (this.error = e));
  }

  redirectToLogin() {
    this.router.navigate(['/']).catch((e) => (this.error = e));
  }

  redirectToProfile() {
    this.router.navigate(['/profile']).catch((e) => (this.error = e));
  }
}
