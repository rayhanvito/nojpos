"use client";

import { useEffect, useState } from 'react';
import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

import type { DashboardSalesPoint } from '@/fixtures/preview/dashboard';

type DashboardSalesChartProps = {
  readonly data: readonly DashboardSalesPoint[];
};

function formatRupiah(value: number) {
  if (value >= 1000000) {
    return `Rp${(value / 1000000).toFixed(value % 1000000 === 0 ? 0 : 1)} jt`;
  }

  if (value >= 1000) {
    return `Rp${Math.round(value / 1000)} rb`;
  }

  return `Rp${value}`;
}

export function DashboardSalesChart({ data }: DashboardSalesChartProps) {
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  if (!isMounted) {
    return <div className="dashboard-chart-frame chart-placeholder" aria-label="Grafik penjualan 7 hari terakhir sedang disiapkan" />;
  }

  return (
    <div className="dashboard-chart-frame" aria-label="Grafik penjualan 7 hari terakhir">
      <ResponsiveContainer width="100%" height="100%" minWidth={0} minHeight={220}>
        <LineChart data={[...data]} margin={{ top: 10, right: 14, left: 0, bottom: 4 }}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} />
          <XAxis dataKey="day" tickLine={false} axisLine={false} />
          <YAxis tickLine={false} axisLine={false} tickFormatter={formatRupiah} width={64} />
          <Tooltip
            formatter={(value) => [formatRupiah(Number(value)), 'Penjualan contoh']}
            labelFormatter={(label) => `Hari ${label}`}
          />
          <Line
            type="monotone"
            dataKey="sales"
            stroke="var(--primary)"
            strokeWidth={3}
            dot={{ r: 4 }}
            activeDot={{ r: 6 }}
            isAnimationActive={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
