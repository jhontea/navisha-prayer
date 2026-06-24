import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  reactStrictMode: true,
  // Produce a self-contained build (.next/standalone) for small Docker images.
  output: 'standalone',
};

export default nextConfig;
