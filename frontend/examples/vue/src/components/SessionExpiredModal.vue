<script setup lang="ts">
import { ref } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();
const modal = ref<HTMLDialogElement>();
const error = ref<Error>();

const show = () => {
  modal.value?.showModal();
};

const redirectToLogin = () => {
  router.push("/").catch((e) => (error.value = e));
};

defineExpose({
  show,
});
</script>

<template>
  <hanko-events @onSessionExpired="show()"></hanko-events>
  <dialog ref="modal">
    <div class="error">{{ error?.message }}</div>
    Please login again.<br /><br />
    <button @click="redirectToLogin()">Login</button>
  </dialog>
</template>
