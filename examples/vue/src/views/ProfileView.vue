<script setup lang="ts">
import HankoProfile from "@/components/HankoProfile.vue";

import { useRouter } from "vue-router";
import { TodoClient } from "@/utils/TodoClient";
import { ref } from "vue";
import type { Ref } from "vue";

const router = useRouter();
const api = import.meta.env.VITE_TODO_API;
const client = new TodoClient(api);
const error: Ref<Error | null> = ref(null);

function setError(e: Error) {
  error.value = e;
}

function todos() {
  router.push("/todo");
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
    <button disabled class="button">Profile</button>
    <button @click.prevent="todos" class="button">Todos</button>
  </nav>
  <main class="content">
    <div class="error">{{ error?.message }}</div>
    <HankoProfile @on-error="setError" />
  </main>
</template>

