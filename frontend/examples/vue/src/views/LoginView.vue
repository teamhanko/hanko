<script setup lang="ts">
import type { Ref } from "vue";
import { ref } from "vue";
import router from "@/router";

const error: Ref<Error | null> = ref(null);

const redirectToTodos = () => {
  router.push("/todo").catch((e) => (error.value = e));
};
</script>

<template>
  <main class="content">
    <div class="error">{{ error?.message }}</div>
    <hanko-auth @onAuthFlowCompleted="redirectToTodos()" />
  </main>
</template>
