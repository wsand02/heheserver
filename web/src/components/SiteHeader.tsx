import React from "react";
import { plainClick } from "../api";

export function SiteHeader({
  theme,
  setTheme,
  visit,
}: {
  theme: string;
  setTheme: (theme: string) => void;
  visit: (url: string) => void;
}) {
  return (
    <header className="site-header">
      <a
        href="/"
        onClick={(e) => {
          if (plainClick(e)) {
            e.preventDefault();
            visit("/");
          }
        }}
      >
        heheserver <span>gallery</span>
      </a>
      <select
        aria-label="Theme"
        value={theme}
        onChange={(e) => setTheme(e.target.value)}
      >
        <option value="auto">System theme</option>
        <option value="light">Light</option>
        <option value="dark">Dark</option>
      </select>
    </header>
  );
}
