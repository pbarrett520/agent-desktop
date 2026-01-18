/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Matrix Theme
        matrix: {
          black: '#0a0a0a',
          darker: '#050505',
          dark: '#0d0d0d',
          panel: '#0f0f0f',
          border: '#1a1a1a',
          green: '#00ff41',
          'green-bright': '#39ff14',
          'green-dim': '#00b336',
          'green-dark': '#004d1a',
          'green-glow': '#00ff4180',
          cyan: '#00ffff',
          'cyan-dim': '#00b3b3',
          red: '#ff3131',
          'red-dim': '#b32222',
          amber: '#ffb000',
          'amber-dim': '#b37b00',
        },
      },
      fontFamily: {
        mono: ['"JetBrains Mono"', '"Fira Code"', '"SF Mono"', 'Monaco', 'Consolas', '"Liberation Mono"', 'monospace'],
        sans: ['"JetBrains Mono"', 'monospace'],
      },
      fontSize: {
        'xs': ['11px', { lineHeight: '1.4' }],
        'sm': ['12px', { lineHeight: '1.5' }],
        'base': ['13px', { lineHeight: '1.6' }],
        'lg': ['15px', { lineHeight: '1.5' }],
        'xl': ['17px', { lineHeight: '1.4' }],
        '2xl': ['20px', { lineHeight: '1.3' }],
        '3xl': ['24px', { lineHeight: '1.2' }],
      },
      boxShadow: {
        'glow': '0 0 10px rgba(0, 255, 65, 0.3)',
        'glow-sm': '0 0 5px rgba(0, 255, 65, 0.2)',
        'glow-lg': '0 0 20px rgba(0, 255, 65, 0.4)',
        'glow-intense': '0 0 30px rgba(0, 255, 65, 0.6), 0 0 60px rgba(0, 255, 65, 0.3)',
        'inner-glow': 'inset 0 0 10px rgba(0, 255, 65, 0.1)',
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'flicker': 'flicker 0.15s infinite',
        'glow-pulse': 'glow-pulse 2s ease-in-out infinite',
        'scan': 'scan 8s linear infinite',
        'typing': 'typing 0.5s steps(1) infinite',
        'matrix-fall': 'matrix-fall 20s linear infinite',
      },
      keyframes: {
        flicker: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.8' },
        },
        'glow-pulse': {
          '0%, 100%': { boxShadow: '0 0 5px rgba(0, 255, 65, 0.2)' },
          '50%': { boxShadow: '0 0 15px rgba(0, 255, 65, 0.4)' },
        },
        scan: {
          '0%': { backgroundPosition: '0 -100vh' },
          '100%': { backgroundPosition: '0 100vh' },
        },
        typing: {
          '0%, 100%': { borderColor: 'transparent' },
          '50%': { borderColor: '#00ff41' },
        },
        'matrix-fall': {
          '0%': { transform: 'translateY(-100%)' },
          '100%': { transform: 'translateY(100vh)' },
        },
      },
      backgroundImage: {
        'scanlines': 'repeating-linear-gradient(0deg, transparent, transparent 1px, rgba(0, 255, 65, 0.03) 1px, rgba(0, 255, 65, 0.03) 2px)',
        'grid': 'linear-gradient(rgba(0, 255, 65, 0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(0, 255, 65, 0.03) 1px, transparent 1px)',
      },
    },
  },
  plugins: [],
}
