import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */

  // Proxy API requests to the Go backend
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
      {
        source: '/health',
        destination: 'http://localhost:8080/health',
      },
    ];
  },
};

export default nextConfig;
