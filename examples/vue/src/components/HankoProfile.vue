<script setup lang="ts">
import { register } from "@teamhanko/hanko-elements";
import { onMounted } from "vue";

const api = import.meta.env.VITE_HANKO_API;
const emit = defineEmits(["on-error"]);

onMounted(() => {
  register({ shadow: true }).catch((e) => emit("on-error", e));
});
</script>

<template>
  <hanko-profile :api="api" />
</template>
