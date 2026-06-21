'use client';

import { useEffect, useState } from 'react';
import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';

type PlatformMiniChartItem = {
  name: string;
  value: number;
};

type PlatformMiniChartProps = {
  title: string;
  description: string;
  data: readonly PlatformMiniChartItem[];
};

export function PlatformMiniChart({ title, description, data }: PlatformMiniChartProps) {
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  return (
    <article className="platform-status-chart">
      <div className="card-header compact">
        <div>
          <span className="card-kicker">Grafik kecil</span>
          <h3>{title}</h3>
        </div>
        <span className="badge info">Preview</span>
      </div>
      <p>{description}</p>
      <div className="platform-chart-frame" aria-label={title}>
        {mounted ? (
          <ResponsiveContainer width="100%" height="100%" minWidth={220} minHeight={180}>
            <BarChart data={data} margin={{ top: 10, right: 8, left: -16, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} />
              <XAxis dataKey="name" tickLine={false} axisLine={false} fontSize={12} />
              <YAxis tickLine={false} axisLine={false} fontSize={12} allowDecimals={false} />
              <Tooltip cursor={{ fill: 'rgba(37, 99, 235, 0.08)' }} />
              <Bar dataKey="value" fill="#2563eb" radius={[8, 8, 0, 0]} isAnimationActive={false} />
            </BarChart>
          </ResponsiveContainer>
        ) : (
          <div className="platform-chart-placeholder">Grafik preview dimuat di browser.</div>
        )}
      </div>
    </article>
  );
}
