import type { ReactNode } from "react";

export function Card({
  title,
  hint,
  actions,
  children,
  className = ""
}: {
  title?: string;
  hint?: string;
  actions?: ReactNode;
  children: ReactNode;
  className?: string;
}) {
  return (
    <section className={`card ${className}`}>
      {(title || actions) && (
        <div className="card-title">
          <div>
            {title}
            {hint && <span className="hint"> {hint}</span>}
          </div>
          {actions}
        </div>
      )}
      {children}
    </section>
  );
}
