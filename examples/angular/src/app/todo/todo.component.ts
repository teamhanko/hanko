import { Component, Input, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { TodoList, TodoService } from '../services/todo.service';

@Component({
  selector: 'app-todo',
  templateUrl: './todo.component.html',
  styleUrls: ['./todo.component.css'],
})
export class TodoComponent implements OnInit {
  todos: TodoList = [];
  error: Error | undefined;
  description = '';

  changeDescription(event: any) {
    this.description = event.target.value;
    console.log(this.description);
  }

  changeCheckbox(event: any) {
    const { currentTarget } = event;
    this.patchTodo(Number(currentTarget.value), currentTarget.checked);
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

  patchTodo(id: number, checked: boolean) {
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
      .then((t) => {
        if (t) {
          this.todos = t;
        }
      })
      .catch((e) => {
        this.error = e;
      });
  }

  deleteTodo(id: number) {
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
      .catch((e) => {
        console.error(e);
      });
  }
}
