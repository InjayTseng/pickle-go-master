const LINE_AUTH_URL = 'https://access.line.me/oauth2/v2.1/authorize';

/**
 * Generate a random string for state parameter
 */
function generateRandomString(length: number): string {
  const array = new Uint8Array(length);
  crypto.getRandomValues(array);
  return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
}

/**
 * Compute HMAC signature for state validation
 * Uses Web Crypto API for secure hashing
 */
async function computeStateHmac(timestamp: string, random: string): Promise<string> {
  // Use a client-side key derived from public config
  // Note: This is not cryptographically secure as the key is public,
  // but provides format validation and time-bounding
  const key = process.env.NEXT_PUBLIC_STATE_KEY || 'pickle-go-state-key';

  const encoder = new TextEncoder();
  const data = encoder.encode(`${timestamp}:${random}`);
  const keyData = encoder.encode(key);

  const cryptoKey = await crypto.subtle.importKey(
    'raw',
    keyData,
    { name: 'HMAC', hash: 'SHA-256' },
    false,
    ['sign']
  );

  const signature = await crypto.subtle.sign('HMAC', cryptoKey, data);
  const hashArray = Array.from(new Uint8Array(signature));
  const base64 = btoa(String.fromCharCode(...hashArray));
  // Return first 16 chars of URL-safe base64
  return base64.replace(/\+/g, '-').replace(/\//g, '_').substring(0, 16);
}

/**
 * Generate a state string for CSRF protection
 * Format: timestamp:random:hmac
 * - timestamp: Unix timestamp for expiry validation
 * - random: Random string for uniqueness
 * - hmac: Signature for integrity validation
 */
export async function generateState(): Promise<string> {
  const timestamp = Math.floor(Date.now() / 1000).toString();
  const random = generateRandomString(8);
  const hmac = await computeStateHmac(timestamp, random);
  return `${timestamp}:${random}:${hmac}`;
}

/**
 * Generate state synchronously (fallback for non-async contexts)
 * Uses simpler format without HMAC
 */
export function generateStateSync(): string {
  const timestamp = Math.floor(Date.now() / 1000).toString();
  const random = generateRandomString(8);
  // For sync version, use a simple hash placeholder
  // This state will still be validated via sessionStorage
  const simpleHash = generateRandomString(8);
  return `${timestamp}:${random}:${simpleHash}`;
}

/**
 * Get the Line Login authorization URL (async version with proper HMAC state)
 */
export async function getLineLoginURL(state?: string): Promise<string> {
  const channelId = process.env.NEXT_PUBLIC_LINE_CHANNEL_ID;
  const redirectUri = process.env.NEXT_PUBLIC_LINE_REDIRECT_URI || `${window.location.origin}/auth/callback`;

  if (!channelId) {
    console.error('LINE_CHANNEL_ID is not configured');
    return '#';
  }

  // Generate state if not provided
  const authState = state || await generateState();

  // Store state in sessionStorage for verification
  if (typeof window !== 'undefined') {
    sessionStorage.setItem('line_auth_state', authState);
  }

  const params = new URLSearchParams({
    response_type: 'code',
    client_id: channelId,
    redirect_uri: redirectUri,
    state: authState,
    scope: 'profile openid',
  });

  return `${LINE_AUTH_URL}?${params.toString()}`;
}

/**
 * Get the Line Login authorization URL (sync version, uses simpler state format)
 */
export function getLineLoginURLSync(state?: string): string {
  const channelId = process.env.NEXT_PUBLIC_LINE_CHANNEL_ID;
  const redirectUri = process.env.NEXT_PUBLIC_LINE_REDIRECT_URI || `${window.location.origin}/auth/callback`;

  if (!channelId) {
    console.error('LINE_CHANNEL_ID is not configured');
    return '#';
  }

  // Generate state if not provided
  const authState = state || generateStateSync();

  // Store state in sessionStorage for verification
  if (typeof window !== 'undefined') {
    sessionStorage.setItem('line_auth_state', authState);
  }

  const params = new URLSearchParams({
    response_type: 'code',
    client_id: channelId,
    redirect_uri: redirectUri,
    state: authState,
    scope: 'profile openid',
  });

  return `${LINE_AUTH_URL}?${params.toString()}`;
}

/**
 * Verify the state parameter from Line callback
 */
export function verifyState(state: string): boolean {
  if (typeof window === 'undefined') {
    return false;
  }

  const storedState = sessionStorage.getItem('line_auth_state');
  sessionStorage.removeItem('line_auth_state');

  return storedState === state;
}

/**
 * Handle the Line Login callback
 * Returns the authorization code if successful
 */
export function handleLineCallback(): { code: string; state: string } | { error: string } {
  if (typeof window === 'undefined') {
    return { error: 'Not in browser environment' };
  }

  const params = new URLSearchParams(window.location.search);
  const code = params.get('code');
  const state = params.get('state');
  const error = params.get('error');
  const errorDescription = params.get('error_description');

  if (error) {
    return { error: errorDescription || error };
  }

  if (!code || !state) {
    return { error: 'Missing code or state parameter' };
  }

  // Verify state
  if (!verifyState(state)) {
    return { error: 'Invalid state parameter' };
  }

  return { code, state };
}
