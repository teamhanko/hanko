<script setup lang="ts">
import { useRouter } from "vue-router";
import { onMounted, ref } from "vue";
import OnSessionExpiredModal from "@/components/SessionExpiredModal.vue";
import { Hanko } from "@teamhanko/hanko-frontend-sdk";

const router = useRouter();

const hankoAPI = import.meta.env.VITE_HANKO_API;
const hankoClient = new Hanko(hankoAPI);

const error = ref<Error>();

const redirectToLogin = () => {
  router.push("/").catch((e) => (error.value = e));
};

const redirectToTodos = () => {
  router.push("/todo").catch((e) => (error.value = e));
};

const logout = () => {
  hankoClient.logout().catch((e) => (error.value = e));
};

onMounted(async () => {
  const {is_valid} = await hankoClient.validateSession();
  if (!is_valid) {
    redirectToLogin();
  }
});
</script>

<template>
  <on-session-expired-modal></on-session-expired-modal>
  <nav class="nav">
    <button @click.prevent="logout" class="button">Logout</button>
    <button disabled class="button">Profile</button>
    <button @click.prevent="redirectToTodos" class="button">Todos</button>
  </nav>
  <main class="content">
    <div class="error">{{ error?.message }}</div>
      <hanko-profile @onUserLoggedOut="redirectToLogin" />
  </main>
</template>
