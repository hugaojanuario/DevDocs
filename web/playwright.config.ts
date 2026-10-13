import { defineConfig } from "@playwright/test";
export default defineConfig({
  testDir: "./e2e",
  fullyParallel: false,
  workers: 1,
  reporter: "list",
  use: {
    baseURL: process.env.DEVDOCS_TEST_URL ?? "http://localhost:8080",
    trace: "off",
    video: "off",
    screenshot: "off",
  },
  timeout: 45_000,
});
