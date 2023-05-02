<script setup lang="ts">
import { register } from "@teamhanko/hanko-elements";
import { onMounted } from "vue";

const api = import.meta.env.VITE_HANKO_API;
const emit = defineEmits(["on-error"]);

onMounted(() => {
  register(api).catch((e) => emit("on-error", e));
});
</script>

<template>
  <hanko-profile />
</template>
