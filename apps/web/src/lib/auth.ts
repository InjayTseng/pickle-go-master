const LINE_AUTH_URL = 'https://access.line.me/oauth2/v2.1/authorize';

/**
 * Generate a random state string for CSRF protection
 */
export function generateState(): string {
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  return Array.from(array, (byte) => byte.toString(16).padStart(2, '0')).join('');
}

/**
 * Get the Line Login authorization URL
 */
export function getLineLoginURL(state?: string): string {
  const channelId = process.env.NEXT_PUBLIC_LINE_CHANNEL_ID;
  const redirectUri = process.env.NEXT_PUBLIC_LINE_REDIRECT_URI || `${window.location.origin}/auth/callback`;

  if (!channelId) {
    console.error('LINE_CHANNEL_ID is not configured');
    return '#';
  }

  // Generate state if not provided
  const authState = state || generateState();

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
