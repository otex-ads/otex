import { useState, useEffect } from "react";

interface Segment {
  text: string;
  className?: string;
}

export function TypewriterText({
  segments,
  speed = 40,
  className = "",
  trigger = "",
}: {
  segments: Segment[];
  speed?: number;
  className?: string;
  trigger?: string | number;
}) {
  const [displayed, setDisplayed] = useState<string[]>([]);
  const [segIndex, setSegIndex] = useState(0);
  const [charIndex, setCharIndex] = useState(0);
  const [done, setDone] = useState(false);

  const reset = () => {
    setDisplayed(segments.map(() => ""));
    setSegIndex(0);
    setCharIndex(0);
    setDone(false);
  };

  useEffect(() => {
    reset();
  }, [trigger]);

  useEffect(() => {
    if (typeof window === "undefined") return;
    const handlePageshow = (e: Event) => {
      if ((e as PageTransitionEvent).persisted) {
        reset();
      }
    };
    window.addEventListener("pageshow", handlePageshow);
    return () => window.removeEventListener("pageshow", handlePageshow);
  }, []);

  useEffect(() => {
    if (segIndex >= segments.length) {
      setDone(true);
      return;
    }

    const seg = segments[segIndex];
    if (charIndex >= seg.text.length) {
      const timeout = setTimeout(() => {
        setSegIndex((i) => i + 1);
        setCharIndex(0);
      }, speed * 3);
      return () => clearTimeout(timeout);
    }

    const timeout = setTimeout(() => {
      setDisplayed((prev) => {
        const next = [...prev];
        next[segIndex] = seg.text.slice(0, charIndex + 1);
        return next;
      });
      setCharIndex((i) => i + 1);
    }, speed);

    return () => clearTimeout(timeout);
  }, [segIndex, charIndex, segments, speed]);

  return (
    <span className={className}>
      {segments.map((seg, i) => (
        <span key={i} className={seg.className}>
          {displayed[i] ?? ""}
        </span>
      ))}
      <span
        className={`inline-block w-[0.07em] h-[0.85em] bg-current ml-0.5 align-middle ${done ? "typewriter-cursor" : ""}`}
        aria-hidden="true"
      />
    </span>
  );
}
