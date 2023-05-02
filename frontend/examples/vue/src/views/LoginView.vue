<script setup lang="ts">
import HankoAuth from "@/components/HankoAuth.vue";
import type { Ref } from "vue";
import { onBeforeUnmount, onMounted, ref } from "vue";
import router from "@/router";
import { createHankoClient } from "@teamhanko/hanko-elements";

const error: Ref<Error | null> = ref(null);

const hankoAPI = import.meta.env.VITE_HANKO_API;
const hankoClient = createHankoClient(hankoAPI);

function redirectToTodos() {
  router.push("/todo");
}

onMounted(() => {
  hankoClient.onAuthFlowCompleted(() => redirectToTodos());
});

onBeforeUnmount(() => {
  hankoClient.removeEventListeners();
});

function setError(e: Error) {
  error.value = e;
}
</script>

<template>
  <main class="content">
    <div class="error">{{ error?.message }}</div>
    <HankoAuth @on-error="setError"/>
  </main>
</template>

