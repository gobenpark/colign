import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { AuthService } from "@/gen/proto/auth/v1/auth_pb";

const transport = createConnectTransport({
  baseUrl: process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080",
});

export const authClient = createClient(AuthService, transport);

const TOKEN_KEY = "colign_access_token";
const REFRESH_KEY = "colign_refresh_token";

export function saveTokens(accessToken: string, refreshToken: string) {
  localStorage.setItem(TOKEN_KEY, accessToken);
  localStorage.setItem(REFRESH_KEY, refreshToken);
}

export function getAccessToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY);
}

export function clearTokens() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_KEY);
}

export function isLoggedIn(): boolean {
  return !!getAccessToken();
}
