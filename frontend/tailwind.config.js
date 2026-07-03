/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Insight Global Theme
        brand: {
          black: '#000000',
          darker: '#0a0a0a',
          dark: '#111111',
          panel: '#151515',
          border: '#262626',
          'border-light': '#333333',
          cyan: '#00D6F2',
          'cyan-dim': '#0099AD',
          'cyan-muted': '#0a3540',
          yellow: '#FDCD01',
          'yellow-dim': '#B39000',
          'yellow-muted': '#3d3202',
          magenta: '#FF0068',
          'magenta-dim': '#B3004A',
          'magenta-muted': '#3d0016',
        },
      },
      fontFamily: {
        sans: ['"Poppins"', '"Inter"', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', '"Fira Code"', 'Consolas', 'monospace'],
      },
      fontSize: {
        'xs': ['11px', { lineHeight: '1.5' }],
        'sm': ['12.5px', { lineHeight: '1.55' }],
        'base': ['13.5px', { lineHeight: '1.6' }],
        'lg': ['15px', { lineHeight: '1.5' }],
        'xl': ['18px', { lineHeight: '1.4' }],
        '2xl': ['21px', { lineHeight: '1.3' }],
        '3xl': ['26px', { lineHeight: '1.2' }],
      },
      boxShadow: {
        'card': '0 1px 2px rgba(0, 0, 0, 0.4)',
        'panel': '0 2px 8px rgba(0, 0, 0, 0.5)',
        'focus-ring': '0 0 0 3px rgba(0, 214, 242, 0.15)',
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      keyframes: {},
      backgroundImage: {},
    },
  },
  plugins: [],
}
