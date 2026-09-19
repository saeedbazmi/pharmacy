"use client";

import dynamic from "next/dynamic";

export const DashboardChart = dynamic(
  () => import("./TrafficChart").then((m) => m.TrafficChart),
  { ssr: false, loading: () => <div className="h-48 rounded-xl bg-border" /> },
);
