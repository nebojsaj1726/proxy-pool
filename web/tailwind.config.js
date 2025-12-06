/** @type {import('tailwindcss').Config} */
export default {
  darkMode: "class",
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        darkBg: "#1f2937",
        darkCard: "#374151",
        darkText: "#f3f4f6",
        alive: "#10b981",
        dead: "#ef4444",
      },
    },
  },
  plugins: [],
};
