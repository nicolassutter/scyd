// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  // spa mode
  ssr: false,
  compatibilityDate: "2025-07-15",
  devtools: { enabled: true },
  modules: ["@nuxt/eslint", "@nuxt/ui", "@vite-pwa/nuxt"],
  css: ["~/assets/css/main.css"],
  devServer: {
    port: 3001,
  },
  pwa: {
    registerType: "autoUpdate",
    includeAssets: ["favicon.ico", "apple-touch-icon.png", "mask-icon.svg"],
    manifest: {
      name: "Scyd",
      short_name: "Scyd",
      description: "A music download manager",
      theme_color: "#1e293b",
      icons: [
        {
          src: "pwa-192x192.png",
          sizes: "192x192",
          type: "image/png",
        },
        {
          src: "pwa-512x512.png",
          sizes: "512x512",
          type: "image/png",
        },
        {
          src: "pwa-512x512.png",
          sizes: "512x512",
          type: "image/png",
          purpose: "any",
        },
        {
          src: "pwa-512x512.png",
          sizes: "512x512",
          type: "image/png",
          purpose: "maskable",
        },
      ],
      share_target: {
        action: "/",
        method: "GET",
        params: {
          url: "url",
          title: "title",
          text: "text",
        },
      },
    },
    registerWebManifestInRouteRules: true,
  },
  typescript: {
    tsConfig: {
      compilerOptions: {
        allowArbitraryExtensions: true,
      },
    },
  },
});
