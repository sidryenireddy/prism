"use client";

import {
  LineChart, Line, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer, Legend,
} from "recharts";
import type { TimeSeriesData } from "@/types";

const COLORS = ["#dc2626", "#171717", "#737373", "#a3a3a3", "#991b1b", "#404040", "#525252"];

interface TimeSeriesRendererProps {
  series: TimeSeriesData[];
  height?: number;
  grouped?: boolean;
}

export function TimeSeriesRenderer({ series, height = 250, grouped }: TimeSeriesRendererProps) {
  if (!series || series.length === 0) {
    return <div className="text-xs text-neutral-400 p-4">No time series data</div>;
  }

  if (!grouped && series.length === 1) {
    const data = series[0].points;
    return (
      <div style={{ width: "100%", height }}>
        <ResponsiveContainer>
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e5e5" />
            <XAxis dataKey="time" tick={{ fontSize: 10 }} />
            <YAxis tick={{ fontSize: 10 }} />
            <Tooltip />
            <Line type="monotone" dataKey="value" stroke="#dc2626" strokeWidth={2} dot={{ r: 2 }} name={series[0].name} />
          </LineChart>
        </ResponsiveContainer>
      </div>
    );
  }

  // Merge all series into unified time points
  const timeSet = new Set<string>();
  series.forEach((s) => s.points.forEach((p) => timeSet.add(p.time)));
  const times = Array.from(timeSet).sort();

  const mergedData = times.map((t) => {
    const point: Record<string, unknown> = { time: t };
    series.forEach((s) => {
      const match = s.points.find((p) => p.time === t);
      point[s.name] = match ? match.value : null;
    });
    return point;
  });

  return (
    <div style={{ width: "100%", height }}>
      <ResponsiveContainer>
        <LineChart data={mergedData}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e5e5" />
          <XAxis dataKey="time" tick={{ fontSize: 10 }} />
          <YAxis tick={{ fontSize: 10 }} />
          <Tooltip />
          <Legend wrapperStyle={{ fontSize: 10 }} />
          {series.map((s, i) => (
            <Line
              key={s.name}
              type="monotone"
              dataKey={s.name}
              stroke={COLORS[i % COLORS.length]}
              strokeWidth={2}
              dot={{ r: 2 }}
              connectNulls
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
