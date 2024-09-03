/** @type {import('tailwindcss').Config} */
export default {
  content: ["./src/**/*.{js,jsx,ts,tsx}"],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/forms")],
  safelist: [
    {
      pattern: /col-start-([1-7])/,
      variants: ["sm"],
    },
    {
      pattern: /text-(blue|pink|gray)-(500|700)/,
      variants: ["group-hover"],
    },
    {
      pattern: /bg-(blue|pink|gray)-(50|100)/,
      variants: ["hover"],
    },
  ]
}

