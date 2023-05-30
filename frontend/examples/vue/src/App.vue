<script setup lang="ts">
import { RouterView } from "vue-router";
import { register } from "@teamhanko/hanko-elements";
import { onMounted } from "vue";

const hankoAPI = import.meta.env.VITE_HANKO_API;

onMounted(() => {
  register(hankoAPI);
});
</script>

<template>
  <RouterView />
</template>

<style>
.content {
  padding: 24px;
  border-radius: 17px;
  color: black;
  background-color: white;
  width: 100%;
  max-width: 500px;
  min-width: 330px;
  margin: 10vh auto;
}

.error {
  color: red;
  padding: 0 0 10px;
}


.button {
  font-size: 1rem;
  border: none;
  background: none;
  cursor: pointer;
}

.button:disabled {
  color: grey !important;
  cursor: default;
  text-decoration: none !important;
}

.nav {
  width: 100%;
  padding: 10px;
  opacity: 0.9;
}


.nav .button:hover {
  text-decoration: underline;
}

.nav .button {
  color: white;
  float: right;
}
</style>
