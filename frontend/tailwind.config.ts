import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        roseMain: '#EF6F9F',
        roseDark: '#C9497A',
        roseSoft: '#FFF0F5',
        rivalBlue: '#70A8E7',
        ink: '#493543',
        muted: '#8D7482'
      },
      boxShadow: {
        soft: '0 10px 30px rgba(143, 74, 105, 0.10)'
      }
    }
  },
  plugins: []
} satisfies Config
