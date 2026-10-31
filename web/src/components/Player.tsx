import React, { useEffect, useRef, useState } from "react";
import type { Item } from "../api";

type FullscreenVideo = HTMLVideoElement & {
  webkitEnterFullscreen?: () => void;
  webkitExitFullscreen?: () => void;
  webkitDisplayingFullscreen?: boolean;
};

function formatTime(seconds: number) {
  const total = Math.max(0, Math.floor(seconds || 0));
  const hours = Math.floor(total / 3600);
  const minutes = Math.floor(total / 60) % 60;
  return `${hours ? `${hours}:` : ""}${String(minutes).padStart(hours ? 2 : 1, "0")}:${String(total % 60).padStart(2, "0")}`;
}

function ControlIcon({
  name,
}: {
  name: "play" | "pause" | "volume" | "muted" | "fullscreen" | "exit";
}) {
  const paths = {
    play: "M8 5v14l11-7Z",
    pause: "M8 5v14M16 5v14",
    volume: "M11 5 6 9H3v6h3l5 4ZM15 8a6 6 0 0 1 0 8M18 5a10 10 0 0 1 0 14",
    muted: "M11 5 6 9H3v6h3l5 4ZM16 9l6 6M22 9l-6 6",
    fullscreen: "M9 3H3v6M15 3h6v6M21 15v6h-6M3 15v6h6",
    exit: "M3 9h6V3M15 3v6h6M21 15h-6v6M9 21v-6H3",
  };
  return (
    <svg
      width="20"
      height="20"
      viewBox="0 0 24 24"
      fill={name === "play" ? "currentColor" : "none"}
      stroke="currentColor"
      strokeWidth="1.8"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d={paths[name]} />
    </svg>
  );
}

function ZoomPreview({ item }: { item: Item }) {
  const image = useRef<HTMLImageElement | HTMLVideoElement>(null);
  const video = useRef<FullscreenVideo>(null);
  const nativeFullscreen = useRef(false);
  const [playing, setPlaying] = useState(false);
  const [duration, setDuration] = useState(0);
  const [time, setTime] = useState(0);
  const [muted, setMuted] = useState(false);
  const [volume, setVolume] = useState(1);
  const [speed, setSpeed] = useState(1);
  const [loop, setLoop] = useState(false);
  const [fullscreen, setFullscreen] = useState(false);
  const togglePlayback = () => {
    const node = video.current;
    if (!node) return;
    if (node.paused) void node.play().catch(() => setPlaying(false));
    else node.pause();
  };
  const seek = (delta: number) => {
    if (video.current && duration)
      video.current.currentTime = Math.max(
        0,
        Math.min(duration, video.current.currentTime + delta),
      );
  };
  useEffect(() => {
    const node = video.current;
    const changed = () =>
      setFullscreen(
        document.fullscreenElement === viewport.current?.parentElement ||
          nativeFullscreen.current ||
          !!node?.webkitDisplayingFullscreen,
      );
    const began = () => {
      nativeFullscreen.current = true;
      changed();
    };
    const ended = () => {
      nativeFullscreen.current = false;
      changed();
    };
    document.addEventListener("fullscreenchange", changed);
    node?.addEventListener("webkitbeginfullscreen", began);
    node?.addEventListener("webkitendfullscreen", ended);
    return () => {
      document.removeEventListener("fullscreenchange", changed);
      node?.removeEventListener("webkitbeginfullscreen", began);
      node?.removeEventListener("webkitendfullscreen", ended);
    };
  }, []);
  const toggleFullscreen = () => {
    const node = video.current;
    const container = viewport.current?.parentElement;
    const enterNative = () => {
      try {
        node?.webkitEnterFullscreen?.();
      } catch {
        // Native fullscreen can also be unavailable or reject the gesture.
      }
    };
    try {
      if (nativeFullscreen.current || node?.webkitDisplayingFullscreen) {
        node?.webkitExitFullscreen?.();
      } else if (document.fullscreenElement) {
        void document.exitFullscreen().catch(() => {});
      } else if (document.fullscreenEnabled && container?.requestFullscreen) {
        try {
          void container.requestFullscreen().catch(enterNative);
        } catch {
          enterNative();
        }
      } else {
        enterNative();
      }
    } catch {
      // A failed fullscreen transition must not interrupt playback.
    }
  };
  const isVideo = item.type === "video";
  const ready = () => {
    const node = image.current;
    return node instanceof HTMLVideoElement
      ? node.videoWidth > 0
      : !!node?.naturalWidth;
  };
  const viewport = useRef<HTMLDivElement>(null);
  const [view, setView] = useState({ scale: 1, x: 0, y: 0 });
  const drag = useRef<{ id: number; x: number; y: number } | null>(null);
  const dragDistance = useRef(0);
  const [dragging, setDragging] = useState(false);
  const [failed, setFailed] = useState(false);
  const reset = () => setView({ scale: 1, x: 0, y: 0 });
  const zoom = (factor: number) => {
    if (!ready()) return;
    setView((old) => {
      const scale = Math.max(1, Math.min(8, old.scale * factor));
      const ratio = scale / old.scale;
      return constrain({ scale, x: old.x * ratio, y: old.y * ratio });
    });
  };
  const constrain = (next: typeof view) => {
    const node = viewport.current,
      img = image.current;
    if (!node || !img) return next;
    const width =
      img instanceof HTMLVideoElement ? img.videoWidth : img.naturalWidth;
    const height =
      img instanceof HTMLVideoElement ? img.videoHeight : img.naturalHeight;
    const fit =
      width && height
        ? Math.min(
            node.clientWidth / width,
            node.clientHeight / height,
            isVideo ? Infinity : 1,
          )
        : 1;
    const x = Math.max(0, (width * fit * next.scale - node.clientWidth) / 2);
    const y = Math.max(0, (height * fit * next.scale - node.clientHeight) / 2);
    return {
      ...next,
      x: Math.max(-x, Math.min(x, next.x)),
      y: Math.max(-y, Math.min(y, next.y)),
    };
  };
  useEffect(() => {
    const node = viewport.current;
    if (!node) return;
    const wheel = (event: WheelEvent) => {
      event.preventDefault();
      if (!ready()) return;
      const rect = node.getBoundingClientRect();
      const x = event.clientX - rect.left - rect.width / 2;
      const y = event.clientY - rect.top - rect.height / 2;
      const delta =
        event.deltaY *
        (event.deltaMode === 1 ? 16 : event.deltaMode === 2 ? rect.height : 1);
      setView((old) => {
        const scale = Math.max(
          1,
          Math.min(8, old.scale * Math.exp(-delta * 0.002)),
        );
        const ratio = scale / old.scale;
        return constrain({
          scale,
          x: x - (x - old.x) * ratio,
          y: y - (y - old.y) * ratio,
        });
      });
    };
    const resize =
      typeof ResizeObserver !== "undefined"
        ? new ResizeObserver(() => setView((old) => constrain(old)))
        : undefined;
    resize?.observe(node);
    node.addEventListener("wheel", wheel, { passive: false });
    return () => {
      node.removeEventListener("wheel", wheel);
      resize?.disconnect();
    };
  }, []);
  useEffect(() => {
    const node = video.current;
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
  return (
    <div
      className={`image-preview${isVideo ? " video-preview" : ""}`}
      data-zoomed={view.scale > 1}
      tabIndex={isVideo ? 0 : undefined}
      aria-label={isVideo ? "Video player" : undefined}
      onKeyDown={(event) => {
        if (
          !isVideo ||
          (event.target as Element).closest("button,input,select") ||
          event.altKey ||
          event.ctrlKey ||
          event.metaKey
        )
          return;
        const key = event.key.toLowerCase();
        if ([" ", "k", "j", "l", "m"].includes(key)) {
          event.preventDefault();
          event.stopPropagation();
        }
        if (key === " " || key === "k") togglePlayback();
        if (key === "j") seek(-10);
        if (key === "l") seek(10);
        if (key === "m" && video.current)
          video.current.muted = !video.current.muted;
      }}
    >
      <div
        className="image-zoom"
        role="group"
        aria-label={isVideo ? "Video zoom" : "Image zoom"}
      >
        <button
          aria-label="Zoom out"
          disabled={view.scale <= 1}
          onClick={() => zoom(1 / 1.25)}
        >
          −
        </button>
        <button onClick={reset} disabled={view.scale === 1}>
          Reset zoom
        </button>
        <button
          aria-label="Zoom in"
          disabled={view.scale >= 8}
          onClick={() => zoom(1.25)}
        >
          +
        </button>
      </div>
      <div
        ref={viewport}
        className={`image-viewport${view.scale > 1 ? " is-zoomed" : ""}${dragging ? " is-dragging" : ""}`}
        onDoubleClick={reset}
        onClick={() => {
          if (isVideo && dragDistance.current < 5) togglePlayback();
        }}
        onPointerDown={(event) => {
          dragDistance.current = 0;
          if (view.scale === 1 || event.button !== 0) return;
          event.preventDefault();
          event.currentTarget.setPointerCapture(event.pointerId);
          drag.current = {
            id: event.pointerId,
            x: event.clientX,
            y: event.clientY,
          };
          setDragging(true);
        }}
        onPointerMove={(event) => {
          const old = drag.current;
          if (!old || old.id !== event.pointerId) return;
          const dx = event.clientX - old.x,
            dy = event.clientY - old.y;
          dragDistance.current += Math.hypot(dx, dy);
          drag.current = {
            id: event.pointerId,
            x: event.clientX,
            y: event.clientY,
          };
          setView((current) =>
            constrain({ ...current, x: current.x + dx, y: current.y + dy }),
          );
        }}
        onLostPointerCapture={() => {
          drag.current = null;
          setDragging(false);
        }}
        onPointerUp={(event) => {
          if (event.currentTarget.hasPointerCapture(event.pointerId))
            event.currentTarget.releasePointerCapture(event.pointerId);
        }}
      >
        {isVideo ? (
          <video
            disableRemotePlayback
            controlsList="noremoteplayback"
            ref={(node) => {
              image.current = node;
              video.current = node;
            }}
            src={item.url}
            poster={item.thumbnail}
            playsInline
            preload="metadata"
            loop={loop}
            aria-label={item.name}
            style={{
              transform: `translate(${view.x}px, ${view.y}px) scale(${view.scale})`,
            }}
            onLoadedMetadata={() =>
              setDuration(
                Number.isFinite(video.current?.duration)
                  ? video.current!.duration
                  : 0,
              )
            }
            onDurationChange={() =>
              setDuration(
                Number.isFinite(video.current?.duration)
                  ? video.current!.duration
                  : 0,
              )
            }
            onTimeUpdate={() => setTime(video.current?.currentTime ?? 0)}
            onPlay={() => setPlaying(true)}
            onPause={() => setPlaying(false)}
            onEnded={() => setPlaying(false)}
            onRateChange={() => setSpeed(video.current?.playbackRate ?? 1)}
            onVolumeChange={() => {
              setMuted(video.current?.muted ?? false);
              setVolume(video.current?.volume ?? 1);
            }}
            onError={() => setFailed(true)}
          />
        ) : (
          <img
            ref={(node) => {
              image.current = node;
            }}
            className="full-image"
            src={item.url}
            alt={item.name}
            draggable={false}
            style={{
              transform: `translate(${view.x}px, ${view.y}px) scale(${view.scale})`,
            }}
            onError={() => setFailed(true)}
          />
        )}
      </div>
      {isVideo && (
        <div
          className="video-controls"
          role="group"
          aria-label="Video playback controls"
        >
          <input
            className="video-seek"
            type="range"
            aria-label="Seek video"
            aria-valuetext={`${formatTime(time)} of ${formatTime(duration)}`}
            min={0}
            max={duration || 0}
            step="0.1"
            value={Math.min(time, duration)}
            disabled={!duration}
            style={
              {
                "--progress": `${duration ? (time / duration) * 100 : 0}%`,
              } as React.CSSProperties
            }
            onChange={(event) => {
              if (video.current)
                video.current.currentTime = Number(event.target.value);
            }}
          />
          <div className="video-control-row">
            <button
              className="video-play"
              aria-label={playing ? "Pause" : "Play"}
              title={playing ? "Pause (K)" : "Play (K)"}
              onClick={togglePlayback}
            >
              <ControlIcon name={playing ? "pause" : "play"} />
            </button>
            <button
              aria-label="Back 10 seconds"
              title="Back 10 seconds (J)"
              disabled={!duration}
              onClick={() => seek(-10)}
            >
              −10<span aria-hidden="true">s</span>
            </button>
            <button
              aria-label="Forward 10 seconds"
              title="Forward 10 seconds (L)"
              disabled={!duration}
              onClick={() => seek(10)}
            >
              +10<span aria-hidden="true">s</span>
            </button>
            <span className="video-time" aria-label="Playback time">
              {formatTime(time)} <span>/ {formatTime(duration)}</span>
            </span>
            <div className="video-control-spacer" />
            <div className="video-volume-group">
              <button
                aria-label={muted || volume === 0 ? "Unmute" : "Mute"}
                title="Toggle sound (M)"
                onClick={() => {
                  if (video.current) {
                    if (video.current.volume === 0) {
                      video.current.volume = 1;
                      video.current.muted = false;
                    } else video.current.muted = !video.current.muted;
                  }
                }}
              >
                <ControlIcon
                  name={muted || volume === 0 ? "muted" : "volume"}
                />
              </button>
              <input
                className="video-volume"
                type="range"
                aria-label="Volume"
                min={0}
                max={1}
                step="0.05"
                value={muted ? 0 : volume}
                onChange={(event) => {
                  if (video.current) {
                    video.current.volume = Number(event.target.value);
                    video.current.muted = false;
                  }
                }}
              />
            </div>
            <select
              aria-label="Playback speed"
              title="Playback speed"
              value={speed}
              onChange={(event) => {
                if (video.current)
                  video.current.playbackRate = Number(event.target.value);
              }}
            >
              {[0.5, 0.75, 1, 1.25, 1.5, 2].map((rate) => (
                <option key={rate} value={rate}>
                  {rate}×
                </option>
              ))}
            </select>
            <button
              aria-label="Loop video"
              aria-pressed={loop}
              title={loop ? "Turn looping off" : "Turn looping on"}
              onClick={() => setLoop((enabled) => !enabled)}
            >
              Loop
            </button>
            <button
              aria-label={fullscreen ? "Exit fullscreen" : "Fullscreen"}
              title={fullscreen ? "Exit fullscreen" : "Fullscreen"}
              onClick={toggleFullscreen}
            >
              <ControlIcon name={fullscreen ? "exit" : "fullscreen"} />
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

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
  if (item.type === "image" || item.type === "video")
    return <ZoomPreview item={item} />;
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
