export interface Todo {
  todoID?: string;
  description: string;
  checked: boolean;
}

export type Todos = Todo[];

export class TodoClient {
  api: string;

  constructor(api: string) {
    this.api = api;
  }

  addTodo(todo: Todo) {
    return fetch(`${this.api}/todo`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(todo),
    });
  }

  listTodos() {
    return fetch(`${this.api}/todo`, {
      credentials: "include",
    });
  }

  patchTodo(id: string, checked: boolean) {
    return fetch(`${this.api}/todo/${id}`, {
      method: "PATCH",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ checked }),
    });
  }

  deleteTodo(id: string) {
    return fetch(`${this.api}/todo/${id}`, {
      method: "DELETE",
      credentials: "include",
    });
  }
}
