"use client";

import { useEffect, useState } from 'react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

import type { DashboardPaymentMethod } from '@/fixtures/preview/dashboard';

type DashboardPaymentChartProps = {
  readonly data: readonly DashboardPaymentMethod[];
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

export function DashboardPaymentChart({ data }: DashboardPaymentChartProps) {
  const [isMounted, setIsMounted] = useState(false);

  useEffect(() => {
    setIsMounted(true);
  }, []);

  if (!isMounted) {
    return <div className="dashboard-chart-frame compact chart-placeholder" aria-label="Grafik metode pembayaran sedang disiapkan" />;
  }

  return (
    <div className="dashboard-chart-frame compact" aria-label="Grafik metode pembayaran">
      <ResponsiveContainer width="100%" height="100%" minWidth={0} minHeight={220}>
        <BarChart data={[...data]} layout="vertical" margin={{ top: 4, right: 14, left: 0, bottom: 4 }}>
          <CartesianGrid strokeDasharray="3 3" horizontal={false} />
          <XAxis type="number" hide />
          <YAxis type="category" dataKey="method" tickLine={false} axisLine={false} width={72} />
          <Tooltip formatter={(value) => [formatRupiah(Number(value)), 'Nominal contoh']} />
          <Bar dataKey="total" fill="var(--primary)" radius={[0, 10, 10, 0]} isAnimationActive={false} />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
