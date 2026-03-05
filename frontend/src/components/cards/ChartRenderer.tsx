"use client";

import {
  BarChart, Bar, LineChart, Line, PieChart, Pie, Cell,
  ScatterChart, Scatter, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer, Legend,
} from "recharts";
import type { CardType, ChartDataPoint } from "@/types";

const COLORS = ["#dc2626", "#171717", "#737373", "#a3a3a3", "#e5e5e5", "#991b1b", "#404040"];

interface ChartRendererProps {
  cardType: CardType;
  data: ChartDataPoint[];
  width?: number;
  height?: number;
}

export function ChartRenderer({ cardType, data, height = 200 }: ChartRendererProps) {
  if (!data || data.length === 0) {
    return <div className="text-xs text-neutral-400 p-4">No data</div>;
  }

  return (
    <div style={{ width: "100%", height }}>
      <ResponsiveContainer>
        {cardType === "bar_chart" ? (
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e5e5" />
            <XAxis dataKey="name" tick={{ fontSize: 10 }} />
            <YAxis tick={{ fontSize: 10 }} />
            <Tooltip />
            <Bar dataKey="value" fill="#dc2626" radius={[2, 2, 0, 0]} />
          </BarChart>
        ) : cardType === "line_chart" ? (
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e5e5" />
            <XAxis dataKey="name" tick={{ fontSize: 10 }} />
            <YAxis tick={{ fontSize: 10 }} />
            <Tooltip />
            <Line type="monotone" dataKey="value" stroke="#dc2626" strokeWidth={2} dot={{ r: 3 }} />
          </LineChart>
        ) : cardType === "pie_chart" ? (
          <PieChart>
            <Pie data={data} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={70} label={({ name }) => name}>
              {data.map((_, i) => (
                <Cell key={i} fill={COLORS[i % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip />
            <Legend wrapperStyle={{ fontSize: 10 }} />
          </PieChart>
        ) : cardType === "scatter_plot" ? (
          <ScatterChart>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e5e5" />
            <XAxis dataKey="name" tick={{ fontSize: 10 }} />
            <YAxis dataKey="value" tick={{ fontSize: 10 }} />
            <Tooltip />
            <Scatter data={data} fill="#dc2626" />
          </ScatterChart>
        ) : cardType === "heat_grid" ? (
          <BarChart data={data} layout="vertical">
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e5e5" />
            <XAxis type="number" tick={{ fontSize: 10 }} />
            <YAxis type="category" dataKey="name" tick={{ fontSize: 10 }} width={80} />
            <Tooltip />
            <Bar dataKey="value" fill="#dc2626" radius={[0, 2, 2, 0]} />
          </BarChart>
        ) : (
          <BarChart data={data}>
            <XAxis dataKey="name" />
            <YAxis />
            <Bar dataKey="value" fill="#dc2626" />
          </BarChart>
        )}
      </ResponsiveContainer>
    </div>
  );
}
