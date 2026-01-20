/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // 啟用獨立輸出模式，用於 Docker 部署
  output: 'standalone',
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'profile.line-scdn.net',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 'storage.googleapis.com',
        pathname: '/**',
      },
    ],
  },
  experimental: {
    serverActions: {
      bodySizeLimit: '2mb',
    },
  },
};

module.exports = nextConfig;
