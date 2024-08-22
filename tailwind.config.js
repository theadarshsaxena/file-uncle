/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./cmd/src/**/*.{html,js}"],
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