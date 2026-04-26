/** @type {import('tailwindcss').Config} */
export default {
  presets: [require("../../../fabrication/src/styles/tailwind.preset.js")],
  content: [
    "./index.html",
    "./src/**/*.{ts,tsx}",
    "../../../fabrication/src/**/*.{ts,tsx}",
  ],
}