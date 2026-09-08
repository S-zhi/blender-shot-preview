/** @type {import('tailwindcss').Config} */
export default {
  darkMode: "class",
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        // OpenHands-inspired dark palette
        background: "#0D0F11",
        surface: {
          100: "#2B313A",
          200: "#21262D",
          300: "#17191C",
          400: "#121417",
          500: "#0D0F11",
        },
        border: {
          subtle: "#30363D",
          default: "#3D444D",
          strong: "#484F58",
        },
        brand: {
          50: "#f0fdfa",
          100: "#ccfbf1",
          500: "#0d9488",
          600: "#0f766e",
          primary: "#14B8A6",
          accent: "#06B6D4",
        },
      },
    },
  },
  plugins: [],
};
