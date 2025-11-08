import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        surface: "#0A1220",
        card: "#0F1B2D",
        "card-hover": "#1A2845",
        primaryBlue: "#5B8CFF",
        primaryTeal: "#22E6B8",
        accentCyan: "#44D6FF",
        accentGreen: "#30E394",
        text: "#E6EEF8",
        muted: "#9BB0C8",
        // Legacy aliases for compatibility
        primary: "#5B8CFF",
        secondary: "#22E6B8",
      },
      borderRadius: {
        card: "20px",
        input: "12px",
        pill: "9999px",
      },
      spacing: {
        "4": "16px",
        "8": "32px",
      },
      boxShadow: {
        xs: "0 1px 2px 0 rgb(0 0 0 / 0.05)",
        sm: "0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)",
        md: "0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)",
        lg: "0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)",
        glow: "0 0 20px rgba(91, 140, 255, 0.3)",
      },
      backdropBlur: {
        md: "12px",
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
      },
      fontWeight: {
        normal: "400",
        semibold: "600",
        bold: "800",
      },
    },
  },
  plugins: [],
};
export default config;
