import { Component, Input, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { Todos, TodoService } from '../services/todo.service';

@Component({
  selector: 'app-todo',
  templateUrl: './todo.component.html',
  styleUrls: ['../app.component.css', './todo.component.css'],
})
export class TodoComponent implements OnInit {
  todos: Todos = [];
  error: Error | undefined;
  description = '';

  changeDescription(event: any) {
    this.description = event.target.value;
  }

  changeCheckbox(event: any) {
    const { currentTarget } = event;
    this.patchTodo(currentTarget.value, currentTarget.checked);
  }

  constructor(private todoService: TodoService, private router: Router) {}

  ngOnInit(): void {
    this.listTodos();
  }

  addTodo(event: any) {
    event.preventDefault();
    const entry = { description: this.description, checked: false };

    this.todoService
      .addTodo(entry)
      .then((res) => {
        if (res.status === 401) {
          this.router.navigate(['/']).catch((e) => (this.error = e));
          return;
        }

        this.description = '';
        this.listTodos();

        return;
      })
      .catch((e) => {
        this.error = e;
      });
  }

  patchTodo(id: string, checked: boolean) {
    this.todoService
      .patchTodo(id, checked)
      .then((res) => {
        if (res.status === 401) {
          this.router.navigate(['/']).catch((e) => (this.error = e));
          return;
        }

        this.listTodos();

        return;
      })
      .catch((e) => {
        this.error = e;
      });
  }

  listTodos() {
    this.todoService
      .listTodos()
      .then((res) => {
        if (res.status === 401) {
          this.router.navigate(['/']).catch((e) => (this.error = e));
          return;
        }

        return res.json();
      })
      .then((todo) => {
        if (todo) {
          this.todos = todo;
        }
      })
      .catch((e) => {
        this.error = e;
      });
  }

  deleteTodo(id: string) {
    this.todoService
      .deleteTodo(id)
      .then((res) => {
        if (res.status === 401) {
          this.router.navigate(['/']).catch((e) => (this.error = e));
          return;
        }

        this.listTodos();

        return;
      })
      .catch((e) => {
        this.error = e;
      });
  }

  logout() {
    this.todoService
      .logout()
      .then(() => {
        this.router.navigate(['/']).catch((e) => (this.error = e));
        return;
      })
      .catch((e) => this.error = e);
  }

  profile() {
    this.router.navigate(['/profile']).catch((e) => (this.error = e));
  }
}
