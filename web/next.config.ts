import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // Do not advertise the framework version to the world.
  poweredByHeader: false,
  // Self-contained server output, so the container image stays small.
  output: "standalone",
};

export default nextConfig;
