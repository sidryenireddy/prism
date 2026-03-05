export type CardType =
  | "filter_object_set"
  | "search_around"
  | "set_math_union"
  | "set_math_intersection"
  | "set_math_difference"
  | "bar_chart"
  | "line_chart"
  | "pie_chart"
  | "scatter_plot"
  | "heat_grid"
  | "object_table"
  | "pivot_table"
  | "count"
  | "sum"
  | "average"
  | "min"
  | "max"
  | "time_series_chart"
  | "rolling_aggregate"
  | "formula"
  | "overlay_chart"
  | "param_object_selection"
  | "param_date_range"
  | "param_numeric"
  | "param_string"
  | "param_boolean"
  | "action_button";

export interface Analysis {
  id: string;
  name: string;
  description: string;
  owner: string;
  share_token?: string;
  created_at: string;
  updated_at: string;
}

export interface Card {
  id: string;
  analysis_id: string;
  card_type: CardType;
  label: string;
  config: Record<string, unknown>;
  position_x: number;
  position_y: number;
  input_card_ids: string[];
  created_at: string;
  updated_at: string;
}

export interface Dashboard {
  id: string;
  analysis_id: string;
  name: string;
  published: boolean;
  layout: DashboardLayout;
  created_at: string;
  updated_at: string;
}

export interface DashboardLayout {
  cards: DashboardCardPlacement[];
}

export interface DashboardCardPlacement {
  card_id: string;
  x: number;
  y: number;
  w: number;
  h: number;
}

export interface CardResult {
  card_id: string;
  card_type: CardType;
  data: unknown;
  error?: string;
}

export interface ExecuteAnalysisResponse {
  results: Record<string, CardResult>;
}

export interface MockObjectType {
  id: string;
  name: string;
  properties: string[];
}

export interface ChartDataPoint {
  name: string;
  value: number;
  [key: string]: unknown;
}

export interface TimeSeriesPoint {
  time: string;
  value: number;
}

export interface TimeSeriesData {
  name: string;
  points: TimeSeriesPoint[];
}

export interface Dataset {
  id: string;
  analysis_id: string;
  card_id: string;
  name: string;
  data: unknown;
  row_count: number;
  created_at: string;
}

export const CARD_CATEGORIES = {
  "Object Set": [
    { type: "filter_object_set" as CardType, label: "Filter Object Set" },
    { type: "search_around" as CardType, label: "Search Around" },
    { type: "set_math_union" as CardType, label: "Union" },
    { type: "set_math_intersection" as CardType, label: "Intersection" },
    { type: "set_math_difference" as CardType, label: "Difference" },
  ],
  Visualization: [
    { type: "bar_chart" as CardType, label: "Bar Chart" },
    { type: "line_chart" as CardType, label: "Line Chart" },
    { type: "pie_chart" as CardType, label: "Pie Chart" },
    { type: "scatter_plot" as CardType, label: "Scatter Plot" },
    { type: "heat_grid" as CardType, label: "Heat Grid" },
  ],
  "Time Series": [
    { type: "time_series_chart" as CardType, label: "Time Series Chart" },
    { type: "rolling_aggregate" as CardType, label: "Rolling Aggregate" },
  ],
  Table: [
    { type: "object_table" as CardType, label: "Object Table" },
    { type: "pivot_table" as CardType, label: "Pivot Table" },
  ],
  Numeric: [
    { type: "count" as CardType, label: "Count" },
    { type: "sum" as CardType, label: "Sum" },
    { type: "average" as CardType, label: "Average" },
    { type: "min" as CardType, label: "Min" },
    { type: "max" as CardType, label: "Max" },
  ],
  Formula: [
    { type: "formula" as CardType, label: "Formula" },
  ],
  Overlay: [
    { type: "overlay_chart" as CardType, label: "Overlay Chart" },
  ],
  Parameter: [
    { type: "param_object_selection" as CardType, label: "Object Selection" },
    { type: "param_date_range" as CardType, label: "Date Range" },
    { type: "param_numeric" as CardType, label: "Numeric" },
    { type: "param_string" as CardType, label: "String" },
    { type: "param_boolean" as CardType, label: "Boolean" },
  ],
  Action: [
    { type: "action_button" as CardType, label: "Action Button" },
  ],
} as const;

export const VIZ_CARD_TYPES: CardType[] = [
  "bar_chart", "line_chart", "pie_chart", "scatter_plot", "heat_grid",
];

export const NUMERIC_CARD_TYPES: CardType[] = [
  "count", "sum", "average", "min", "max",
];

export const TABLE_CARD_TYPES: CardType[] = [
  "object_table", "pivot_table",
];

export const PARAM_CARD_TYPES: CardType[] = [
  "param_object_selection", "param_date_range", "param_numeric", "param_string", "param_boolean",
];

export const TIME_SERIES_CARD_TYPES: CardType[] = [
  "time_series_chart", "rolling_aggregate",
];
