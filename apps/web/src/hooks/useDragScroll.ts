"use client";

import { useRef, useEffect, useCallback } from "react";

export function useDragScroll() {
  const ref = useRef<HTMLDivElement>(null);
  const isDown = useRef(false);
  const startX = useRef(0);
  const scrollLeft = useRef(0);

  const handleMouseDown = useCallback((e: MouseEvent) => {
    const el = ref.current;
    if (!el) return;
    isDown.current = true;
    el.classList.add("cursor-grabbing");
    el.classList.remove("cursor-grab");
    startX.current = e.pageX - el.offsetLeft;
    scrollLeft.current = el.scrollLeft;
  }, []);

  const handleMouseLeave = useCallback(() => {
    const el = ref.current;
    if (!el) return;
    isDown.current = false;
    el.classList.remove("cursor-grabbing");
    el.classList.add("cursor-grab");
  }, []);

  const handleMouseUp = useCallback(() => {
    const el = ref.current;
    if (!el) return;
    isDown.current = false;
    el.classList.remove("cursor-grabbing");
    el.classList.add("cursor-grab");
  }, []);

  const handleMouseMove = useCallback((e: MouseEvent) => {
    const el = ref.current;
    if (!el || !isDown.current) return;
    e.preventDefault();
    const x = e.pageX - el.offsetLeft;
    const walk = (x - startX.current) * 1.5; // scroll speed multiplier
    el.scrollLeft = scrollLeft.current - walk;
  }, []);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.classList.add("cursor-grab");
    el.addEventListener("mousedown", handleMouseDown);
    el.addEventListener("mouseleave", handleMouseLeave);
    el.addEventListener("mouseup", handleMouseUp);
    el.addEventListener("mousemove", handleMouseMove);
    return () => {
      el.removeEventListener("mousedown", handleMouseDown);
      el.removeEventListener("mouseleave", handleMouseLeave);
      el.removeEventListener("mouseup", handleMouseUp);
      el.removeEventListener("mousemove", handleMouseMove);
    };
  }, [handleMouseDown, handleMouseLeave, handleMouseUp, handleMouseMove]);

  return ref;
}
