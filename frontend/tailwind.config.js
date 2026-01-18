/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Primary
        primary: {
          blue: '#298FC2',
          DEFAULT: '#298FC2',
        },
        // Neutral
        neutral: {
          gray: '#696158',
          light: '#CCCCCB',
          DEFAULT: '#696158',
        },
        // Secondary
        secondary: {
          lightBlue: '#A3D4EC',
          ember: '#EEB927',
          navy: '#01405C',
          purple: '#7D87C2',
          coral: '#E34154',
          orange: '#F89848',
          lime: '#BCC883',
        },
        // Link
        link: {
          DEFAULT: '#298FC2',
        },
      },
      fontFamily: {
        sans: ["Arial", "Helvetica", "system-ui", "-apple-system", "Segoe UI", "Roboto", "sans-serif"],
      },
      fontWeight: {
        regular: '400',
        bold: '700',
        black: '900',
      },
      fontSize: {
        body: ['16px', { lineHeight: '1.45' }],
        h6: ['18px', { lineHeight: '1.2' }],
        h5: ['20px', { lineHeight: '1.2' }],
        h4: ['24px', { lineHeight: '1.2' }],
        h3: ['28px', { lineHeight: '1.2' }],
        h2: ['32px', { lineHeight: '1.2' }],
        h1: ['36px', { lineHeight: '1.2' }],
      },
      lineHeight: {
        tight: '1.2',
        normal: '1.45',
      },
      borderRadius: {
        sm: '4px',
        md: '8px',
      },
      boxShadow: {
        card: '0 1px 3px rgba(0,0,0,0.08)',
      },
    },
  },
  plugins: [],
}
