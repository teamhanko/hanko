<script setup>
import { useRouter } from "vue-router";
import { register } from "@teamhanko/hanko-elements";
import { onMounted } from "vue";

onMounted(() => {
  register({ shadow: true }).catch((e) => console.error(e));
});

const api = import.meta.env.VITE_HANKO_API;

const router = useRouter();

const redirectToTodo = () => {
  router.push({ path: "/todo" });
};
</script>

<template>
  <h1>Hello from { testString }</h1>
  <hanko-auth @hankoAuthSuccess="redirectToTodo" :api="api" />
</template>
