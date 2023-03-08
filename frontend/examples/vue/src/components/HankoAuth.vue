<script setup lang="ts">
import { useRouter } from "vue-router";
import { register } from "@teamhanko/hanko-elements";
import { onMounted } from "vue";

const router = useRouter();
const api = import.meta.env.VITE_HANKO_API;
const emit = defineEmits(["on-error"]);

const redirectToTodo = () => {
  router.push({ path: "/todo" });
};

onMounted(() => {
  register({ shadow: true }).catch((e) => emit("on-error", e));
});
</script>

<template>
  <hanko-auth @hankoAuthSuccess="redirectToTodo" :api="api" />
</template>
