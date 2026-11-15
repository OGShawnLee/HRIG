import { defineConfig, presetWind3, presetWebFonts, transformerDirectives } from "unocss"

export default defineConfig({
  presets: [
    presetWind3(),
    presetWebFonts({
      provider: "fontshare",
      fonts: {
        sans: "Satoshi",
      },
    }),
  ],
  transformers: [
    transformerDirectives({
      applyVariable: "--uno",
    }),
  ],
  theme: {
    colors: {
      main: "black",
      side: "#404040",
      dull: "#737373",
      ground: "#FAFAFA",
      light: "#F5F5F5",
    },
  },
  shortcuts: {
    button: "h-12 min-w-fit px-4 border border-black border-solid rounded-none inline-flex items-center justify-center disabled:opacity-50",
    "button-main": "button bg-main font-medium text-white",
    "button-side": "button bg-ground text-main",
    field: "h-12 px-4 border border-solid border-dull rounded-none font-medium placeholder-dull disabled:opacity-50",
  },
})
