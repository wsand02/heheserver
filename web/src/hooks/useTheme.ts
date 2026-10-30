import { useEffect, useState } from "react";

export function useTheme() {
  const [theme, setTheme] = useState(() => {
    try {
      return localStorage.getItem("theme") || "auto";
    } catch {
      return "auto";
    }
  });
  useEffect(() => {
    const preference = matchMedia("(prefers-color-scheme: dark)");
    const update = () => {
      document.documentElement.dataset.theme =
        theme === "auto" ? (preference.matches ? "dark" : "light") : theme;
    };
    update();
    preference.addEventListener("change", update);
    try {
      localStorage.setItem("theme", theme);
    } catch {}
    return () => preference.removeEventListener("change", update);
  }, [theme]);
  return [theme, setTheme] as const;
}
