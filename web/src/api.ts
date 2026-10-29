import type React from "react";

export type Item = {
  name: string;
  path: string;
  type: string;
  size: number;
  modified: string;
  width: number;
  height: number;
  url: string;
  thumbnail?: string;
  viewerUrl: string;
};
export type Gallery = {
  items: Item[];
  path: string;
  page: number;
  pages: number;
  total: number;
  infinite: boolean;
  version: string;
};
export type Media = {
  item: Item;
  previous?: string;
  next?: string;
  position: number;
  total: number;
  galleryUrl: string;
};
export async function getJSON<T>(url: string, signal: AbortSignal): Promise<T> {
  const response = await fetch(url, { signal });
  const data = await response.json();
  if (!response.ok) throw new Error(data.error?.message || "Request failed");
  return data;
}
export function apiURL(kind: string, url: string) {
  return "/api/" + kind + new URL(url, location.origin).search;
}
export function galleryURL(url: string) {
  const p = new URL(url, location.origin).searchParams;
  const path = p.get("path") || "/";
  p.set("path", path.slice(0, path.lastIndexOf("/") + 1) || "/");
  return "/?" + p;
}
export function plainClick(e: React.MouseEvent) {
  return e.button === 0 && !e.metaKey && !e.ctrlKey && !e.shiftKey && !e.altKey;
}
