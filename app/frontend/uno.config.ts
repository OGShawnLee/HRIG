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
      ground: "#000000",
      surface: "#0A0A0A",
      main: "#FFFFFF",
      side: "#D4D4D8",
      dull: "#9CA3AF",
      line: "#FFFFFF",
      lineDim: "#262626",
      ok: "#4ADE80",
      warn: "#FACC15",
      crit: "#F87171",
      stale: "#9CA3AF",
    },
  },
  shortcuts: {
    "hud-card": "bg-ground border border-solid border-line rounded-2xl",
    "hud-pill": "bg-ground border border-solid border-line rounded-xl",
    "hud-value": "font-medium tabular-nums",
    button:
      "h-11 min-w-fit px-4 border border-solid border-line rounded-xl inline-flex items-center justify-center transition-colors disabled:opacity-50",
    "button-main": "button bg-main font-medium text-ground hover:bg-side",
    "button-side": "button bg-ground text-main hover:bg-surface",
    field:
      "h-11 px-4 bg-ground text-main border border-solid border-line rounded-xl font-medium placeholder-dull focus:outline-none disabled:opacity-50",
  },
})
