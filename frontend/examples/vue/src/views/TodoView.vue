<script setup lang="ts">
import type { Todos } from "@/utils/TodoClient";
import { TodoClient } from "@/utils/TodoClient";
import { useRouter } from "vue-router";
import type { Ref } from "vue";
import { onMounted, ref } from "vue";

const router = useRouter();

const api = import.meta.env.VITE_TODO_API;
const client = new TodoClient(api);

const error: Ref<Error | null> = ref(null);
const todos: Ref<Todos> = ref([]);
const description = ref("");

onMounted(() => {
  listTodos();
});

function changeDescription(event: any) {
  description.value = event.currentTarget.value;
}

const changeCheckbox = (event: any) => {
  const { currentTarget } = event;
  patchTodo(currentTarget.value, currentTarget.checked);
};

function addTodo() {
  const todo = { description: description.value, checked: false };

  client
    .addTodo(todo)
    .then((res) => {
      if (res.status === 401) {
        router.push("/");
        return;
      }

      description.value = "";
      listTodos();

      return;
    })
    .catch((e) => {
      error.value = e;
    });
}

function listTodos() {
  client
    .listTodos()
    .then((res) => {
      if (res.status === 401) {
        router.push("/");
        return;
      }

      return res.json();
    })
    .then((todo) => {
      if (todo) {
        todos.value = todo;
      }
    })
    .catch((e) => {
      error.value = e;
    });
}

const patchTodo = (id: string, checked: boolean) => {
  client
    .patchTodo(id, checked)
    .then((res) => {
      if (res.status === 401) {
        router.push("/");
        return;
      }

      listTodos();

      return;
    })
    .catch((e) => {
      error.value = e;
    });
};

const deleteTodo = (id: string) => {
  client
    .deleteTodo(id)
    .then((res) => {
      if (res.status === 401) {
        router.push("/");
        return;
      }

      listTodos();

      return;
    })
    .catch((e) => {
      error.value = e;
    });
};

function profile() {
  router.push("/profile");
}

function logout() {
  client
    .logout()
    .then(() => {
      router.push("/");
      return;
    })
    .catch((e) => {
      error.value = e;
    });
}
</script>

<template>
  <nav class="nav">
    <button @click.prevent="logout" class="button">Logout</button>
    <button @click.prevent="profile" class="button">Profile</button>
    <button disabled class="button">Todos</button>
  </nav>
  <div class="content">
    <h1 class="headline">Todos</h1>
    <div class="error">{{ error?.message }}</div>
    <form @submit.prevent="addTodo" class="form">
      <input
        required
        class="input"
        type="text"
        :value="description"
        @change="changeDescription"
      />
      <button type="submit" class="button">+</button>
    </form>
    <div class="list">
      <div v-for="(todo, index) in todos" class="item" :key="index">
        <input
          class="checkbox"
          :id="todo.todoID"
          type="checkbox"
          :value="todo.todoID"
          :checked="todo.checked"
          @change="changeCheckbox"
        />
        <label class="description" :for="todo.todoID">{{
          todo.description
        }}</label>
        <button class="button" @click="() => deleteTodo(todo.todoID)">×</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.nav {
  width: 100%;
  padding: 10px;
  opacity: 0.9;
}

.button {
  font-size: 1rem;
  border: none;
  background: none;
  cursor: pointer;
}

.button:disabled {
  color: grey !important;
  cursor: default;
  text-decoration: none !important;
}

.nav .button:hover {
  text-decoration: underline;
}

.nav .button {
  color: white;
  float: right;
}

.content {
  padding: 24px;
  border-radius: 17px;
  color: black;
  background-color: white;
  width: 100%;
  max-width: 500px;
  min-width: 330px;
  margin: 10vh auto;
}

.headline {
  text-align: center;
  margin-top: 0;
}

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

.error {
  color: red;
  padding: 0 0 10px;
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
