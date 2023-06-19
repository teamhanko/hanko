<script setup lang="ts">
import { ref } from "vue";
import router from "@/router";

const error = ref<Error>();

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
