import { defineConfig } from "$fresh/server.ts";
import twindPlugin from "$fresh/plugins/twind.ts"
import twindConfig from "./twind.config.ts";
export default defineConfig({
  port: 8888,
  plugins: [twindPlugin(twindConfig)],
});
