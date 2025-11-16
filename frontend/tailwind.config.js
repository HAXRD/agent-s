/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // 深色主题颜色
        dark: {
          bg: '#1e1e1e',
          surface: '#2d2d2d',
          text: '#d1d5db',
        },
      },
    },
  },
  plugins: [],
}

