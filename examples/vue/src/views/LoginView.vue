<script setup lang="ts">
import HankoAuth from "@/components/HankoAuth.vue";
</script>

<template>
  <main class="content">
    <HankoAuth />
  </main>
</template>

<style scoped>
.content {
  padding: 24px;
  border-radius: 17px;
  color: black;
  background-color: white;
  width: 500px;
  position: fixed;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}
</style>
