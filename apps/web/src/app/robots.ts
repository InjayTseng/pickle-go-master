import { MetadataRoute } from 'next';

const BASE_URL = process.env.NEXT_PUBLIC_BASE_URL || 'https://picklego.tw';

/**
 * Generate robots.txt for SEO
 * @see https://nextjs.org/docs/app/api-reference/file-conventions/metadata/robots
 */
export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: '*',
        allow: [
          '/',
          '/events/',
          '/g/',
        ],
        disallow: [
          '/my/',
          '/my/*',
          '/api/',
          '/auth/',
          '/_next/',
          '/private/',
        ],
      },
      // Special rules for common bots
      {
        userAgent: 'Googlebot',
        allow: [
          '/',
          '/events/',
          '/g/',
        ],
        disallow: [
          '/my/',
          '/api/',
          '/auth/',
        ],
      },
      // Line crawler
      {
        userAgent: 'Line',
        allow: [
          '/',
          '/events/',
          '/g/',
        ],
        disallow: [
          '/my/',
          '/api/',
        ],
      },
    ],
    sitemap: `${BASE_URL}/sitemap.xml`,
    host: BASE_URL,
  };
}
