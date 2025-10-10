import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: {
    path: "http://localhost:3000/openapi.json",
  },
  output: "app/utils/client",
});
