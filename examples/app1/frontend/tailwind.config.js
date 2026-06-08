/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx,ts,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        serif: ['Lora', 'ui-serif', 'Georgia', 'serif'],
      },
      colors: {
        paper: {
          50: '#fbf9f3',
          100: '#f5f1e6',
          200: '#ece5d2',
          300: '#dccfa8',
        },
        ink: {
          900: '#1d1b16',
          700: '#3d3a31',
          500: '#6a655a',
          400: '#8b8678',
        },
      },
    },
  },
  plugins: [],
}
