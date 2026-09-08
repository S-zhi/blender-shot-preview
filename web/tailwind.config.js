/** @type {import('tailwindcss').Config} */
export default {
  darkMode: "class",
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        // ── OpenHands-aligned Surface palette ──
        surface: {
          DEFAULT: "#050505",
          card: "#0a0a0a",
          elevated: "#1a1a1a",
          outline: "#171717",
          background: "#262626",
          divider: "#525252",
          button: "#737373",
          text: "#A3A3A3",
        },
        // ── Border system ──
        border: {
          DEFAULT: "#242424",
          hover: "#3a3a3a",
        },
        // ── Content / text colors ──
        content: {
          DEFAULT: "#fafafa",
          muted: "#8c8c8c",
          icon: "#3a3a3a",
        },
        // ── Semantic status colors ──
        status: {
          "success-bg": "rgba(16, 185, 129, 0.1)",
          "success-border": "rgba(16, 185, 129, 0.4)",
          "success-text": "#6ee7b7",
          "success-badge-bg": "rgba(16, 185, 129, 0.15)",
          "fail-bg": "rgba(244, 63, 94, 0.1)",
          "fail-border": "rgba(244, 63, 94, 0.4)",
          "fail-text": "#fda4af",
          "fail-solid": "#dc2626",
        },
        // ── Brand / accent (project identity teal) ──
        brand: {
          50: "#f0fdfa",
          100: "#ccfbf1",
          500: "#0d9488",
          600: "#0f766e",
          primary: "#14B8A6",
          accent: "#06B6D4",
        },
        // ── Overlay & pill ──
        "muted-overlay": "rgba(5, 5, 5, 0.4)",
        "pill-bg": "rgba(31, 31, 31, 0.3)",
      },
      fontFamily: {
        sans: [
          "Outfit",
          "ui-sans-serif",
          "system-ui",
          "-apple-system",
          "BlinkMacSystemFont",
          '"Segoe UI"',
          "Roboto",
          '"Helvetica Neue"',
          "Arial",
          "sans-serif",
        ],
        mono: [
          '"IBM Plex Mono"',
          "ui-monospace",
          "SFMono-Regular",
          "Menlo",
          "Monaco",
          "Consolas",
          '"Courier New"',
          "monospace",
        ],
      },
      keyframes: {
        "fade-in": {
          "0%": { opacity: "0", transform: "translateY(4px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        "slide-in-left": {
          "0%": { opacity: "0", transform: "translateX(-12px)" },
          "100%": { opacity: "1", transform: "translateX(0)" },
        },
        "modal-enter": {
          "0%": { opacity: "0", transform: "scale(0.96)" },
          "100%": { opacity: "1", transform: "scale(1)" },
        },
      },
      animation: {
        "fade-in": "fade-in 0.25s ease-out",
        "slide-in-left": "slide-in-left 0.2s ease-out",
        "modal-enter": "modal-enter 0.2s ease-out",
      },
    },
  },
  plugins: [],
};
