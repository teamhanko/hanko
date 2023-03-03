import { Injectable } from '@angular/core';
import { environment } from '../../environments/environment';

export interface Todo {
  todoID?: string;
  description: string;
  checked: boolean;
}

export type Todos = Todo[];

@Injectable({
  providedIn: 'root',
})
export class TodoService {
  api = 'http://localhost:8002';

  constructor() {
    this.api = environment.todoApi;
  }

  addTodo(todo: Todo) {
    return fetch(`${this.api}/todo`, {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(todo),
    });
  }

  listTodos() {
    return fetch(`${this.api}/todo`, {
      credentials: 'include',
    });
  }

  patchTodo(id: string, checked: boolean) {
    return fetch(`${this.api}/todo/${id}`, {
      method: 'PATCH',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({checked}),
    });
  }

  deleteTodo(id: string) {
    return fetch(`${this.api}/todo/${id}`, {
      method: 'DELETE',
      credentials: 'include',
    });
  }

  logout() {
    return fetch(`${this.api}/logout`, {
      credentials: 'include',
    });
  }
}
