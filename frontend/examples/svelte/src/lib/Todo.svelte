<script lang="ts">
  import { onMount } from "svelte";
  import { useNavigate } from "svelte-navigator";
  import type { Todos } from "./TodoClient"
  import { TodoClient } from "./TodoClient";

  const api = import.meta.env.VITE_TODO_API;
  const todoClient = new TodoClient(api);

  const navigate = useNavigate();

  let description = '';
  let error: Error | null = null;
  let todos: Todos = [];

  onMount(async () => {
    listTodos();
  });

  const changeDescription = (event: any) => {
    description = event.currentTarget.value;
  };

  const changeCheckbox = (event: any) => {
    const { currentTarget } = event;
    patchTodo(currentTarget.value, currentTarget.checked);
  };

  const addTodo = () => {
    const entry = { description: description, checked: false };

    todoClient
      .addTodo(entry)
      .then((res) => {
        if (res.status === 401) {
          navigate("/");
          return;
        }

        description = "";
        listTodos();

        return;
      })
      .catch((e) => {
        error = e;
      });
  }

  const listTodos = () => {
    todoClient
      .listTodos()
      .then((res) => {
        if (res.status === 401) {
          navigate("/");
          return;
        }

        return res.json();
      })
      .then((todo) => {
        if (todo) {
          todos = todo;
        }
      })
      .catch((e) => {
        error = e;
      });
  }

  const patchTodo = (id: string, checked: boolean) => {
    todoClient
      .patchTodo(id, checked)
      .then((res) => {
        if (res.status === 401) {
          navigate("/");
          return;
        }

        listTodos();

        return;
      })
      .catch((e) => {
        error = e;
      });
  };

  const deleteTodo = (id: string) => {
    todoClient
      .deleteTodo(id)
      .then((res) => {
        if (res.status === 401) {
          navigate("/");
          return;
        }

        listTodos();

        return;
      })
      .catch((e) => {
        error = e;
      });
  };

  const logout = () => {
    todoClient
      .logout()
      .then(() => {
        navigate("/");
        return;
      })
      .catch((e) => {
        console.error(e);
      });
  }

  const profile = () => {
    navigate("/profile");
  }
</script>

<nav class="nav">
  <button class="button" on:click|preventDefault={logout}>Logout</button>
  <button class="button" on:click|preventDefault={profile}>Profile</button>
  <button class="button" disabled>Todos</button>
</nav>
<div class="content">
  <h1 class="headline">Todos</h1>
  {#if error}
    <div class="error">{ error?.message }</div>
  {/if}
  <form on:submit|preventDefault={addTodo} class="form">
    <input
      required
      class="input"
      type="text"
      value={description}
      on:change={changeDescription}
    />
    <button type="submit" class="button">+</button>
  </form>
  <div class="list">
    {#each todos as todo}
      <div class="item">
        <input
          class="checkbox"
          type="checkbox"
          id={todo.todoID}
          value={todo.todoID}
          checked={todo.checked}
          on:change={changeCheckbox}
        />
        <label class="description" for={todo.todoID}>{todo.description}</label>
        <button class="button" on:click={() => deleteTodo(todo.todoID)}>×</button>
      </div>
    {/each}
  </div>
</div>


<style>
  .form {
      display: flex;
      margin-bottom: 17px;
  }

  .form .input {
      flex-grow: 1;
      margin-right: 10px;
  }

  .form .button {
      color: black;
  }

  .list {
      display: flex;
      flex-direction: column;
      row-gap: 7px;
  }

  .item {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      column-gap: 7px;
  }

  .description {
      flex-grow: 1;
      cursor: pointer;
  }

  .input {
      border: 1px solid black;
      border-radius: 2.4px;
  }

  .checkbox {
      margin-left: 0;
      -webkit-appearance: none;
      appearance: none;
      background-color: #fff;
      font: inherit;
      color: currentColor;
      width: 1em;
      height: 1em;
      border: 1px solid currentColor;
      border-radius: 0.15em;
      transform: translateY(-0.075em);
      display: grid;
      place-content: center;
  }

  .checkbox::before {
      content: "";
      width: 0.65em;
      height: 0.65em;
      transform: scale(0);
      box-shadow: inset 1em 1em black;

      transform-origin: bottom left;
      clip-path: polygon(14% 44%, 0 65%, 50% 100%, 100% 16%, 80% 0%, 43% 62%);
  }

  .checkbox:checked::before {
      transform: scale(1);
  }
</style>
