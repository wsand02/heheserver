import { defineConfig } from "vite";
export default defineConfig({
  build: { outDir: "../internal/frontend/dist", emptyOutDir: true },
  server: {
    proxy: {
      "/api": "http://localhost:3400",
      "/fs": "http://localhost:3400",
      "/resize": "http://localhost:3400",
      "/vidthumb": "http://localhost:3400",
    },
  },
});
