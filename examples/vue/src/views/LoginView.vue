<script setup lang="ts">
import HankoAuth from "@/components/HankoAuth.vue";
import { ref } from "vue";
import type { Ref } from "vue";

const error: Ref<Error | null> = ref(null);

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

