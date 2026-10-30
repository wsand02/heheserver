import React, { useEffect, useRef } from "react";

export function Filters({
  url,
  apply,
  clear,
}: {
  url: string;
  apply: (form: HTMLFormElement) => void;
  clear: () => void;
}) {
  const form = useRef<HTMLFormElement>(null);
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  useEffect(() => {
    const p = new URL(url, location.origin).searchParams;
    const node = form.current!;
    for (const key of ["q", "ext"]) {
      (node.elements.namedItem(key) as HTMLInputElement).value =
        p.get(key) || "";
    }
    node
      .querySelectorAll<HTMLInputElement>('[name="type"]')
      .forEach((input) => {
        input.checked = p.getAll("type").includes(input.value);
      });
  }, [url]);
  useEffect(() => () => clearTimeout(timer.current), []);
  function submit() {
    clearTimeout(timer.current);
    if (form.current) apply(form.current);
  }
  return (
    <form
      ref={form}
      className="filters"
      onInput={() => {
        clearTimeout(timer.current);
        timer.current = setTimeout(submit, 250);
      }}
      onSubmit={(e) => {
        e.preventDefault();
        submit();
      }}
    >
      <label>
        Search
        <input name="q" placeholder="Find a filename" />
      </label>
      <details>
        <summary>Filters</summary>
        <fieldset>
          <legend>File types</legend>
          {["image", "video", "audio", "dir", "other"].map((type) => (
            <label key={type}>
              <input type="checkbox" name="type" value={type} />
              {type}
            </label>
          ))}
        </fieldset>
        <label>
          Extensions
          <input name="ext" placeholder=".png, .mp4" />
        </label>
      </details>
      <button type="submit" hidden />
      <button
        type="button"
        onClick={() => {
          clearTimeout(timer.current);
          clear();
        }}
      >
        Clear
      </button>
    </form>
  );
}
