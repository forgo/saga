import { defineConfig } from "vite";
import { fresh } from "@fresh/plugin-vite";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [fresh(), tailwindcss()],
  server: {
    // Allow all origins for dev server (fixes virtual module CORS issues)
    cors: true,
    // Configure HMR to use proper origin
    hmr: {
      protocol: "ws",
      host: "localhost",
    },
  },
  // Optimize deps to handle Fresh virtual modules
  optimizeDeps: {
    exclude: ["fresh"],
  },
});
