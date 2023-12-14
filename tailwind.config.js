// tailwind.config.js
let plugin = require("tailwindcss/plugin");


/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/pages/template/**/*.tmpl"],
  theme: {
    colors : {
      'primary': '#de1823',
      'secondary': '#f67200',
      'mainbg': '#090502',
      'white': 'white',
      'gold' : '#edb712',
      'silver' : '#dfdeea',
      'dark': '#150603'
    },
    extend: {},
  },
  plugins: [
    plugin(function ({ matchVariant, theme }) {
      matchVariant(
        'nth',
        (value) => {
          return `&:nth-child(${value})`;
        },
        {
          values: {
            DEFAULT: 'n', // Default value for `nth:`
            '2n': '2n', // `nth-2n:utility` will generate `:nth-child(2n)` CSS selector
            '3n': '3n',
            '4n': '4n',
            '5n': '5n',
            //... so on if you need
          },
        }
      );
    }),
  ],
}

