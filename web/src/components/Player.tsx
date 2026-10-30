import React, { useEffect, useRef, useState } from "react";
import type { Item } from "../api";

export function Player({ item }: { item: Item }) {
  const media = useRef<HTMLMediaElement | null>(null);
  const [failed, setFailed] = useState(false);
  useEffect(() => {
    const node = media.current;
    return () => {
      if (node) {
        node.pause();
        node.removeAttribute("src");
        node.load();
      }
    };
  }, []);
  if (failed)
    return (
      <p role="alert">
        Preview unavailable. <a href={item.url}>Open original</a>
      </p>
    );
  if (item.type === "image")
    return (
      <img
        className="full-image"
        src={item.url}
        alt={item.name}
        onError={() => setFailed(true)}
      />
    );
  if (item.type === "video")
    return (
      <video
        ref={(node) => {
          media.current = node;
        }}
        src={item.url}
        controls
        playsInline
        preload="metadata"
        onError={() => setFailed(true)}
      />
    );
  if (item.type === "audio")
    return (
      <audio
        ref={(node) => {
          media.current = node;
        }}
        src={item.url}
        controls
        preload="metadata"
        onError={() => setFailed(true)}
      />
    );
  return (
    <p>This file has no preview. Use the download or open-original action.</p>
  );
}
