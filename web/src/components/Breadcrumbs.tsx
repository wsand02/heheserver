import React from "react";
import { plainClick } from "../api";

export function Breadcrumbs({
  parts,
  visit,
}: {
  parts: string[];
  visit: (url: string) => void;
}) {
  return (
    <nav aria-label="Breadcrumb" className="breadcrumbs">
      <a
        href="/"
        onClick={(e) => {
          if (plainClick(e)) {
            e.preventDefault();
            visit("/");
          }
        }}
      >
        Home
      </a>
      {parts.map((part, i) => {
        const u =
          "/?" +
          new URLSearchParams({
            path: "/" + parts.slice(0, i + 1).join("/") + "/",
          });
        return (
          <React.Fragment key={u}>
            <span>/</span>
            <a
              href={u}
              onClick={(e) => {
                if (plainClick(e)) {
                  e.preventDefault();
                  visit(u);
                }
              }}
            >
              {part}
            </a>
          </React.Fragment>
        );
      })}
    </nav>
  );
}
