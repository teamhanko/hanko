export interface Todo {
  todoID?: string;
  description: string;
  checked: boolean;
}

export type Todos = Todo[];

export class TodoClient {
  api: string;
  bearerToken?: string;

  constructor(api: string, bearerToken?: string) {
    this.api = api;
    this.bearerToken = bearerToken;
  }

  addTodo(todo: Todo) {
    return fetch(`${this.api}/todo`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${this.bearerToken}`,
      },
      body: JSON.stringify(todo),
    });
  }

  listTodos() {
    return fetch(`${this.api}/todo`, {
      credentials: "include",
      headers: {
        "Authorization": `Bearer ${this.bearerToken}`,
      }
    });
  }

  patchTodo(id: string, checked: boolean) {
    return fetch(`${this.api}/todo/${id}`, {
      method: "PATCH",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${this.bearerToken}`,
      },
      body: JSON.stringify({ checked }),
    });
  }

  deleteTodo(id: string) {
    return fetch(`${this.api}/todo/${id}`, {
      method: "DELETE",
      credentials: "include",
      headers: {
        "Authorization": `Bearer ${this.bearerToken}`,
      }
    });
  }

  logout() {
    return fetch(`${this.api}/logout`, {
      credentials: "include",
      headers: {
        "Authorization": `Bearer ${this.bearerToken}`,
      }
    });
  }
}

