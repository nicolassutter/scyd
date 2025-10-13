import { client } from "~/utils/client/client.gen";

export default defineNuxtRouteMiddleware(async () => {
  // Set the base URL for the API client before making any requests
  client.setConfig({
    baseUrl: import.meta.env.DEV ? "http://localhost:3000" : undefined,
    throwOnError: true,
    credentials: import.meta.env.DEV ? "include" : "same-origin",
  });
});
