import React from "react";
import clsx from "clsx";
import { twMerge } from "tailwind-merge";

interface BadgeProps {
  variant?: "success" | "warning" | "info" | "neutral";
  children: React.ReactNode;
  className?: string;
}

export const Badge: React.FC<BadgeProps> = ({
  variant = "neutral",
  children,
  className,
}) => {
  const base =
    "inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium border";

  const variants = {
    success: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    warning: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    info: "bg-cyan-500/10 text-cyan-400 border-cyan-500/20",
    neutral: "bg-surface-200 text-gray-300 border-border-subtle",
  };

  return (
    <span className={twMerge(clsx(base, variants[variant], className))}>
      {children}
    </span>
  );
};
