<script setup lang="ts">
import { TodoClient } from "@/utils/TodoClient";
import type { TodoList } from "@/utils/TodoClient";
import { useRouter } from "vue-router";
import { onMounted, ref } from "vue";
import type { Ref } from "vue";

const router = useRouter();

const api = import.meta.env.VITE_TODO_API;
const client = new TodoClient(api);

const error: Ref<Error | null> = ref(null);
const todos: Ref<TodoList> = ref([]);
const description = ref("");

onMounted(() => {
  listTodos();
});

function changeDescription(event: any) {
  description.value = event.currentTarget.value;
}

const changeCheckbox = (event: any) => {
  const { currentTarget } = event;
  patchTodo(Number(currentTarget.value), currentTarget.checked);
};

function addTodo() {
  const entry = { description: description.value, checked: false };

  client
    .addTodo(entry)
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
    .then((t) => {
      if (t) {
        todos.value = t;
      }
    })
    .catch((e) => {
      error.value = e;
    });
}

const patchTodo = (id: number, checked: boolean) => {
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

const deleteTodo = (id: number) => {
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
    <button @click="logout" class="button">logout</button>
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
      <div v-for="(todo, id) in todos" class="item" :key="id">
        <input
          class="checkbox"
          :id="id.toString(10)"
          type="checkbox"
          :value="id"
          :checked="todo.checked"
          @change="changeCheckbox"
        />
        <label class="description" :for="id.toString(10)">{{
          todo.description
        }}</label>
        <button class="button" @click="() => deleteTodo(id)">×</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.nav {
  width: 100%;
  position: fixed;
  top: 0;
  padding: 10px;
  opacity: 0.9;
}

.button {
  font-size: 1rem;
  border: none;
  background: none;
  cursor: pointer;
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
  width: 500px;
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
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
  transition: 120ms transform ease-in-out;
  box-shadow: inset 1em 1em black;

  transform-origin: bottom left;
  clip-path: polygon(14% 44%, 0 65%, 50% 100%, 100% 16%, 80% 0%, 43% 62%);
}

.checkbox:checked::before {
  transform: scale(1);
}
</style>
