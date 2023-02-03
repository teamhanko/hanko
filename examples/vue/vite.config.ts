import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

// https://vitejs.dev/config/
export default defineConfig({
  optimizeDeps: {
    include: ['@teamhanko/hanko-elements'],
  },
  build: {
    commonjsOptions: { include: ['@teamhanko/hanko-elements'] },
  },
  plugins: [
    vue({
      template: {
        compilerOptions: { isCustomElement: (tag) => tag === "hanko-auth" },
      },
    }),
  ],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
});
