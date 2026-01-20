import { ImageResponse } from 'next/og';
import { NextRequest } from 'next/server';

export const runtime = 'edge';

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);

    // Get parameters from URL
    const title = searchParams.get('title') || 'Pickle Go';
    const subtitle = searchParams.get('subtitle') || '匹克球揪團平台';
    const location = searchParams.get('location') || '';
    const date = searchParams.get('date') || '';
    const spots = searchParams.get('spots') || '';

    return new ImageResponse(
      (
        <div
          style={{
            height: '100%',
            width: '100%',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: '#ffffff',
            backgroundImage: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
          }}
        >
          {/* Main Card */}
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              backgroundColor: 'white',
              borderRadius: '24px',
              padding: '60px 80px',
              boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
              maxWidth: '900px',
            }}
          >
            {/* Logo */}
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                marginBottom: '20px',
              }}
            >
              <span
                style={{
                  fontSize: '36px',
                  fontWeight: 'bold',
                  color: '#667eea',
                }}
              >
                Pickle Go
              </span>
            </div>

            {/* Title */}
            <div
              style={{
                fontSize: '48px',
                fontWeight: 'bold',
                color: '#1f2937',
                textAlign: 'center',
                marginBottom: '16px',
                maxWidth: '700px',
              }}
            >
              {title}
            </div>

            {/* Subtitle/Location */}
            {location && (
              <div
                style={{
                  fontSize: '28px',
                  color: '#6b7280',
                  textAlign: 'center',
                  marginBottom: '8px',
                }}
              >
                {location}
              </div>
            )}

            {/* Date & Spots */}
            <div
              style={{
                display: 'flex',
                gap: '24px',
                marginTop: '20px',
              }}
            >
              {date && (
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    backgroundColor: '#f3f4f6',
                    padding: '12px 24px',
                    borderRadius: '12px',
                    fontSize: '24px',
                    color: '#374151',
                  }}
                >
                  {date}
                </div>
              )}
              {spots && (
                <div
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    backgroundColor: '#dcfce7',
                    padding: '12px 24px',
                    borderRadius: '12px',
                    fontSize: '24px',
                    color: '#166534',
                  }}
                >
                  {spots}
                </div>
              )}
            </div>

            {/* Subtitle */}
            <div
              style={{
                fontSize: '24px',
                color: '#9ca3af',
                marginTop: '24px',
              }}
            >
              {subtitle}
            </div>
          </div>
        </div>
      ),
      {
        width: 1200,
        height: 630,
      }
    );
  } catch (error) {
    console.error('Error generating OG image:', error);
    return new Response('Failed to generate image', { status: 500 });
  }
}
