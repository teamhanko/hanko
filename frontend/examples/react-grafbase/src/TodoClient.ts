export interface Todo {
  todoID?: string;
  title: string;
  checked: boolean;
  complete: boolean;
}

export type Todos = Todo[];

export class TodoClient {
  api: string;

  constructor(api: string) {
    this.api = api;
  }

  getCookie(name: string): string | null {
    const nameLenPlus = (name.length + 1);
    return document.cookie
      .split(';')
      .map(c => c.trim())
      .filter(cookie => {
        return cookie.substring(0, nameLenPlus) === `${name}=`;
      })
      .map(cookie => {
        return decodeURIComponent(cookie.substring(nameLenPlus));
      })[0] || null;
  }

  addTodo(todo: Todo) {
    const CreateTodoMutation = `
    mutation TodoCreate($title: String!, $complete: Boolean!) {
      todoCreate(input: {title: $title, complete: $complete}){
        todo {
          complete
          id
          title
        }
      }
    }
    `

    const response = fetch(`${this.api}/graphql`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + this.getCookie("hanko"),
      },
      body: JSON.stringify({
        query: CreateTodoMutation,
        variables: {
          title: todo.title,
          complete: todo.complete
        }
      })
    })
    return response
  }


  listTodos() {
    const GetAllTodosQuery = /* GraphQL */ `
      query TodoCollection {
        todoCollection(first: 10) {
          edges {
            node {
              complete
              id
              title
            }
          }
        }
      }
    `

    const response = fetch(`${this.api}/graphql`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + this.getCookie("hanko"),
      },
      body: JSON.stringify({
        query: GetAllTodosQuery,
        variables: {
          first: 100
        }
      })
    })
    return response
  }

  patchTodo(id: string, checked: boolean) {
    const UpdateTodoMutation = `
          mutation UpdateTodoById($id: ID!, $newComplete: Boolean!) {
          todoUpdate(by: { id: $id }, input: { complete: $newComplete }) {
          todo {
            id
            title
            complete
          }
        }
      }
    `

    const response = fetch(`${this.api}/graphql`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + this.getCookie("hanko"),
      },
      body: JSON.stringify({
        query: UpdateTodoMutation,
        variables: {
          id: id,
          newComplete: checked
        }
      })
    })
    return response
  }

  deleteTodo(id: string) {
    const DeleteTodoMutation = `
      mutation TodoDelete($id: ID!) {
        todoDelete(by: {id: $id}){
          deletedId
        }
      }`

    const response = fetch(`${this.api}/graphql`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + this.getCookie("hanko"),
      },
      body: JSON.stringify({
        query: DeleteTodoMutation,
        variables: {
          id: id,
        }
      })
    })
    return response
  }
}
