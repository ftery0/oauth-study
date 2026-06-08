/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'ui-monospace', 'SFMono-Regular', 'monospace'],
      },
      boxShadow: {
        card: '0 1px 2px rgba(15,23,42,0.06), 0 1px 3px rgba(15,23,42,0.08)',
        cardHover: '0 4px 12px rgba(79,70,229,0.18), 0 2px 4px rgba(15,23,42,0.08)',
      },
    },
  },
  plugins: [],
}
