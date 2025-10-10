// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  // spa mode
  ssr: false,
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  modules: ["@nuxt/eslint", "@nuxt/ui"],
  css: ["~/assets/css/main.css"],
});
