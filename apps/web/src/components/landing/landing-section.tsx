import type { ReactNode } from 'react';

type LandingSectionProps = {
  id?: string;
  eyebrow?: string;
  title: string;
  description?: string;
  children: ReactNode;
  className?: string;
};

export function LandingSection({ id, eyebrow, title, description, children, className }: LandingSectionProps) {
  return (
    <section className={['landing-section', className].filter(Boolean).join(' ')} id={id}>
      <div className="landing-container">
        <div className="landing-section-header">
          {eyebrow ? <p className="landing-eyebrow">{eyebrow}</p> : null}
          <h2>{title}</h2>
          {description ? <p>{description}</p> : null}
        </div>
        {children}
      </div>
    </section>
  );
}
