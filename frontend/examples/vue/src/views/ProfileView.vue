<script setup lang="ts">
import { useRouter } from "vue-router";
import type { Ref } from "vue";
import { ref } from "vue";
import OnSessionExpiredModal from "@/components/SessionExpiredModal.vue";
import { useHanko } from "@/composables/useHanko";

const router = useRouter();
const { hankoClient } = useHanko();
const error: Ref<Error | null> = ref(null);

const redirectToLogin = () => {
  router.push("/").catch((e) => (error.value = e));
};

const redirectToTodos = () => {
  router.push("/todo").catch((e) => (error.value = e));
};

const logout = () => {
  hankoClient.user.logout().catch((e) => (error.value = e));
};
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
    <hanko-profile
      @onSessionNotPresent="redirectToLogin"
      @onUserLoggedOut="redirectToLogin"
    />
  </main>
</template>
