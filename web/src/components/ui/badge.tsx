import React from "react";
import clsx from "clsx";
import { twMerge } from "tailwind-merge";

interface BadgeProps {
  variant?: "success" | "warning" | "danger" | "info" | "neutral";
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
    success: "bg-status-success-bg text-status-success-text border-status-success-border",
    warning: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    danger: "bg-status-fail-bg text-status-fail-text border-status-fail-border",
    info: "bg-brand-primary/10 text-brand-primary border-brand-primary/20",
    neutral: "bg-surface-elevated text-content-muted border-border",
  };

  return (
    <span className={twMerge(clsx(base, variants[variant], className))}>
      {children}
    </span>
  );
};
