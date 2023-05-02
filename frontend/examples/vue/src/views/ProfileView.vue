<script setup lang="ts">
import HankoProfile from "@/components/HankoProfile.vue";

import { useRouter } from "vue-router";
import type { Ref } from "vue";
import { onBeforeUnmount, onMounted, ref } from "vue";
import { createHankoClient } from "@teamhanko/hanko-elements";

const router = useRouter();
const hankoAPI = import.meta.env.VITE_HANKO_API;
const hankoClient = createHankoClient(hankoAPI);
const error: Ref<Error | null> = ref(null);

onMounted(() => {
  hankoClient.onSessionRemoved(() => redirectToLogin());
});

onBeforeUnmount(() => {
  hankoClient.removeEventListeners();
});

function setError(e: Error) {
  error.value = e;
}

function redirectToLogin() {
  console.log("profile redirectToLogin");
  router.push("/");
}

function redirectToTodos() {
  router.push("/todo");
}

function logout() {
  hankoClient.user.logout().catch((e) => {
    error.value = e;
  });
}
</script>

<template>
  <nav class="nav">
    <button @click.prevent="logout" class="button">Logout</button>
    <button disabled class="button">Profile</button>
    <button @click.prevent="redirectToTodos" class="button">Todos</button>
  </nav>
  <main class="content">
    <div class="error">{{ error?.message }}</div>
    <HankoProfile @on-error="setError" />
  </main>
</template>
