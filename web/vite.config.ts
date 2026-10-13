import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        configure(proxy) {
          proxy.on("proxyReq", (req, source) => {
            if (
              source.headers.origin === "http://localhost:5173" ||
              source.headers.origin === "http://127.0.0.1:5173"
            )
              req.setHeader("Origin", "http://localhost:8080");
          });
        },
      },
    },
  },
  build: { sourcemap: false },
});
