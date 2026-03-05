"use client";

import {
  ComposedChart, Bar, Line, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer, Legend,
} from "recharts";

const COLORS = ["#dc2626", "#171717", "#737373", "#a3a3a3", "#991b1b"];

interface OverlayLayer {
  cardId: string;
  type: string;
  chartData?: { name: string; value: number }[];
  series?: { name: string; points: { time: string; value: number }[] }[];
}

interface OverlayChartRendererProps {
  layers: OverlayLayer[];
  height?: number;
}

export function OverlayChartRenderer({ layers, height = 250 }: OverlayChartRendererProps) {
  if (!layers || layers.length === 0) {
    return <div className="text-xs text-neutral-400 p-4">Connect chart cards as inputs to overlay them</div>;
  }

  // Merge data from all layers into a unified dataset
  const mergedMap = new Map<string, Record<string, number>>();
  const dataKeys: { key: string; type: string; color: string }[] = [];

  layers.forEach((layer, layerIdx) => {
    const key = `layer_${layerIdx}`;
    const color = COLORS[layerIdx % COLORS.length];

    if (layer.chartData) {
      dataKeys.push({ key, type: layer.type || "bar", color });
      layer.chartData.forEach((d) => {
        const existing = mergedMap.get(d.name) || {};
        existing[key] = d.value;
        mergedMap.set(d.name, existing);
      });
    } else if (layer.series && layer.series.length > 0) {
      dataKeys.push({ key, type: "line", color });
      layer.series[0].points.forEach((p) => {
        const existing = mergedMap.get(p.time) || {};
        existing[key] = p.value;
        mergedMap.set(p.time, existing);
      });
    }
  });

  const names = Array.from(mergedMap.keys()).sort();
  const data = names.map((name) => ({
    name,
    ...mergedMap.get(name),
  }));

  if (data.length === 0) {
    return <div className="text-xs text-neutral-400 p-4">No overlay data</div>;
  }

  return (
    <div style={{ width: "100%", height }}>
      <ResponsiveContainer>
        <ComposedChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e5e5" />
          <XAxis dataKey="name" tick={{ fontSize: 10 }} />
          <YAxis yAxisId="left" tick={{ fontSize: 10 }} />
          {dataKeys.length > 1 && (
            <YAxis yAxisId="right" orientation="right" tick={{ fontSize: 10 }} />
          )}
          <Tooltip />
          <Legend wrapperStyle={{ fontSize: 10 }} />
          {dataKeys.map((dk, i) =>
            dk.type === "line" ? (
              <Line
                key={dk.key}
                yAxisId={i === 0 ? "left" : "right"}
                type="monotone"
                dataKey={dk.key}
                stroke={dk.color}
                strokeWidth={2}
                dot={{ r: 2 }}
                name={`Series ${i + 1}`}
              />
            ) : (
              <Bar
                key={dk.key}
                yAxisId={i === 0 ? "left" : "right"}
                dataKey={dk.key}
                fill={dk.color}
                radius={[2, 2, 0, 0]}
                name={`Series ${i + 1}`}
              />
            )
          )}
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
}
