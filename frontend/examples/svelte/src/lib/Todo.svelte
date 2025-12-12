<script lang="ts">
  import { onMount } from "svelte";
  import { useNavigate } from "svelte-navigator";
  import type { Todos } from "./TodoClient"
  import { TodoClient } from "./TodoClient";
  import { Hanko } from "@teamhanko/hanko-elements";
  import SessionExpiredModal from "./SessionExpiredModal.svelte";

  const todoAPI = import.meta.env.VITE_TODO_API;
  const todoClient = new TodoClient(todoAPI);

  const hankoAPI = import.meta.env.VITE_HANKO_API;
  const hankoClient = new Hanko(hankoAPI);

  const navigate = useNavigate();

  let openSessionExpiredModal = $state();
  let description = $state('');
  let error: Error | null = $state(null);
  let todos: Todos = $state([]);

  const changeDescription = (event: any) => {
    description = event.currentTarget.value;
  };

  const changeCheckbox = (event: any) => {
    const { currentTarget } = event;
    patchTodo(currentTarget.value, currentTarget.checked);
  };

  const addTodo = (event) => {
    event.preventDefault();
    const entry = { description: description, checked: false };

    todoClient
      .addTodo(entry)
      .then((res) => {
        if (res.status === 401) {
          //showModal = true;
          openSessionExpiredModal();
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
          openSessionExpiredModal();
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
          openSessionExpiredModal();
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
          openSessionExpiredModal();
          return;
        }

        listTodos();

        return;
      })
      .catch((e) => {
        error = e;
      });
  };

  const logout = (event) => {
    event.preventDefault();
    hankoClient.logout()
      .catch((e) => {
        error = e;
      });
  }

  const redirectToLogin = () => {
    navigate("/");
  }

  const redirectToProfile = (event) => {
    event.preventDefault();
    navigate("/profile");
  }

  onMount(async () => {
    const {is_valid} = await hankoClient.validateSession();
    if (is_valid) {
      listTodos();
    } else {
      redirectToLogin();
    }
  });
</script>
<SessionExpiredModal bind:openSessionExpiredModal></SessionExpiredModal>
<hanko-events ononUserLoggedOut={redirectToLogin}></hanko-events>
<nav class="nav">
  <button class="button" onclick={logout}>Logout</button>
  <button class="button" onclick={redirectToProfile}>Profile</button>
  <button class="button" disabled>Todos</button>
</nav>
<div class="content">
  <h1 class="headline">Todos</h1>
  {#if error}
    <div class="error">{ error?.message }</div>
  {/if}
  <form onsubmit={addTodo} class="form">
    <input
      required
      class="input"
      type="text"
      value={description}
      onchange={changeDescription}
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
          onchange={changeCheckbox}
        />
        <label class="description" for={todo.todoID}>{todo.description}</label>
        <button class="button" onclick={() => deleteTodo(todo.todoID)}>×</button>
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
