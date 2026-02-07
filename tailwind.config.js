/** @type {import('tailwindcss').Config} */
const path = require('path');

module.exports = {
    content: [
        path.resolve(__dirname, './internal/src/**/*.{html,js,jsx,ts,tsx}'),
        path.resolve(__dirname, './cmd/**/*.{html,js,jsx,ts,tsx}')
    ],
    theme: {
      extend: {
        fontFamily: {
          custom: ['Urbanist', 'sans-serif'],
        },
        maxHeight: {
          'screen-40': '40vh', // 40% of the viewport height
        },
      },
    },
    plugins: [],
  }