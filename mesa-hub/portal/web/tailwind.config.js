/** @type {import('tailwindcss').Config} */
export default {
  presets: [require("../../../fabrication/skin/src/styles/tailwind.preset.js")],
  content: [
    "./index.html",
    "./src/**/*.{ts,tsx}",
    "../../../fabrication/skin/src/**/*.{ts,tsx}",
  ],
}
