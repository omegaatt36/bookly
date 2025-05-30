/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Nord Theme palette
        nord: {
          // Polar Night
          0: '#2e3440',
          1: '#3b4252',
          2: '#434c5e',
          3: '#4c566a',
          // Snow Storm
          4: '#d8dee9',
          5: '#e5e9f0',
          6: '#eceff4',
          // Frost
          7: '#8fbcbb',
          8: '#88c0d0',
          9: '#81a1c1',
          10: '#5e81ac',
          // Aurora
          11: '#bf616a', // red
          12: '#d08770', // orange
          13: '#ebcb8b', // yellow
          14: '#a3be8c', // green
          15: '#b48ead', // purple
        },
        // Semantic color mapping
        background: {
          DEFAULT: '#2e3440',
          secondary: '#3b4252',
          tertiary: '#434c5e',
          accent: '#4c566a',
        },
        foreground: {
          DEFAULT: '#d8dee9',
          secondary: '#e5e9f0',
          accent: '#eceff4',
        },
        primary: {
          DEFAULT: '#88c0d0',
          foreground: '#2e3440',
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#88c0d0',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
        },
        secondary: {
          DEFAULT: '#81a1c1',
          foreground: '#2e3440',
        },
        accent: {
          DEFAULT: '#5e81ac',
          foreground: '#eceff4',
        },
        destructive: {
          DEFAULT: '#bf616a',
          foreground: '#eceff4',
        },
        success: {
          DEFAULT: '#a3be8c',
          foreground: '#2e3440',
        },
        warning: {
          DEFAULT: '#ebcb8b',
          foreground: '#2e3440',
        },
        info: {
          DEFAULT: '#b48ead',
          foreground: '#eceff4',
        },
        muted: {
          DEFAULT: '#434c5e',
          foreground: '#d8dee9',
        },
        border: '#4c566a',
        input: '#3b4252',
        ring: '#88c0d0',
      },
      borderRadius: {
        lg: '0.5rem',
        md: '0.375rem',
        sm: '0.25rem',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      animation: {
        'fade-in': 'fadeIn 0.2s ease-in-out',
        'slide-in': 'slideIn 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideIn: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
      },
    },
  },
  plugins: [],
}
