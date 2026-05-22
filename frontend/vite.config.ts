import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      "/auth": "http://localhost:8080",
      "/users": "http://localhost:8080",
      "/generate": "http://localhost:8080",
    },
  },
});
