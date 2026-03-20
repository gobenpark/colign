"use client";

import { useState, useEffect, useCallback } from "react";
import { usePathname, useRouter } from "next/navigation";
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "./app-sidebar";
import { authClient, clearTokens, getAccessToken } from "@/lib/auth";

const NO_SIDEBAR_PATHS = ["/auth", "/onboarding"];

export function SidebarLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const [hasToken, setHasToken] = useState(false);
  const [verified, setVerified] = useState(false);

  const verifySession = useCallback(async () => {
    const token = getAccessToken();
    if (!token) {
      setHasToken(false);
      setVerified(true);
      return;
    }

    setHasToken(true);

    // Skip verification on auth pages
    if (NO_SIDEBAR_PATHS.some((p) => pathname.startsWith(p))) {
      setVerified(true);
      return;
    }

    try {
      await authClient.me({}, {
        headers: { Authorization: `Bearer ${token}` },
      });
      setVerified(true);
    } catch {
      // Token invalid or user doesn't exist
      clearTokens();
      setHasToken(false);
      setVerified(true);
      router.replace("/auth");
    }
  }, [pathname, router]);

  // Verify on mount and pathname change
  useEffect(() => {
    verifySession();
  }, [verifySession]);

  // Verify on window focus (tab switch back)
  useEffect(() => {
    const handleFocus = () => {
      verifySession();
    };
    window.addEventListener("focus", handleFocus);
    return () => window.removeEventListener("focus", handleFocus);
  }, [verifySession]);

  const hideSidebar = NO_SIDEBAR_PATHS.some((p) => pathname.startsWith(p)) || !hasToken;

  // Show nothing until verified to prevent flash
  if (!verified) {
    return null;
  }

  if (hideSidebar) {
    return <>{children}</>;
  }

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        {/* Mobile hamburger */}
        <div className="flex items-center gap-2 border-b border-border/50 px-4 py-2 md:hidden">
          <SidebarTrigger />
          <span className="text-sm font-semibold">Colign</span>
        </div>
        {children}
      </SidebarInset>
    </SidebarProvider>
  );
}
