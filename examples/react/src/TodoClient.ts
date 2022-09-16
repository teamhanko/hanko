export interface TodoEntry {
  description: string;
  checked: boolean;
}

export type TodoList = TodoEntry[];

export class TodoClient {
  api: string;

  constructor(api: string) {
    this.api = api;
  }

  async addTodo(t: TodoEntry) {
    return await fetch(`${this.api}/todo`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(t),
    });
  }

  async listTodos() {
    return await fetch(`${this.api}/todo`, {
      credentials: "include",
    });
  }

  async patchTodo(id: number, checked: boolean) {
    return await fetch(`${this.api}/todo/${id}`, {
      method: "PATCH",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ checked }),
    });
  }

  async deleteTodo(id: number) {
    return await fetch(`${this.api}/todo/${id}`, {
      method: "DELETE",
      credentials: "include",
    });
  }

  async logout() {
    return await fetch(`${this.api}/logout`, {
      credentials: "include",
    });
  }
}
