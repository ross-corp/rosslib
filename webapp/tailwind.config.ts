import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ["var(--font-inter)", "ui-sans-serif", "system-ui", "sans-serif"],
        mono: ["var(--font-jetbrains-mono)", "ui-monospace", "monospace"],
      },
      colors: {
        surface: {
          0: "var(--surface-0)",
          1: "var(--surface-1)",
          2: "var(--surface-2)",
          3: "var(--surface-3)",
        },
        border: {
          DEFAULT: "var(--border)",
          strong: "var(--border-strong)",
        },
        text: {
          primary: "var(--text-primary)",
          secondary: "var(--text-secondary)",
          tertiary: "var(--text-tertiary)",
          inverted: "var(--text-inverted)",
        },
        accent: {
          DEFAULT: "var(--accent)",
          hover: "var(--accent-hover)",
        },
        semantic: {
          error: {
            DEFAULT: "var(--error-text)",
            bg: "var(--error-bg)",
            border: "var(--error-border)",
          },
          success: {
            DEFAULT: "var(--success-text)",
            bg: "var(--success-bg)",
            border: "var(--success-border)",
          },
          warning: {
            DEFAULT: "var(--warning-text)",
            bg: "var(--warning-bg)",
            border: "var(--warning-border)",
          },
          info: {
            DEFAULT: "var(--info-text)",
            bg: "var(--info-bg)",
            border: "var(--info-border)",
          },
        },
        danger: {
          DEFAULT: "var(--danger-btn-text)",
          bg: "var(--danger-btn-bg)",
          "bg-hover": "var(--danger-btn-bg-hover)",
        },
        link: "var(--link-color)",
        like: {
          DEFAULT: "var(--like-color)",
          hover: "var(--like-color-hover)",
        },
        badge: {
          DEFAULT: "var(--badge-bg)",
          text: "var(--badge-text)",
        },
        toggle: {
          bg: "var(--toggle-bg)",
          active: "var(--toggle-bg-active)",
          knob: "var(--toggle-knob)",
        },
        progress: "var(--progress-bar)",
      },
      backgroundColor: {
        overlay: "var(--overlay)",
      },
      maxWidth: {
        shell: "1280px",
      },
      keyframes: {
        "toast-in": {
          "0%": { opacity: "0", transform: "translateY(8px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
      },
      animation: {
        "toast-in": "toast-in 0.2s ease-out",
      },
    },
  },
  plugins: [],
};

export default config;
