"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { orgClient } from "@/lib/organization";
import { useOrg } from "@/lib/org-context";

export default function OnboardingPage() {
  const router = useRouter();
  const { currentOrg, loading: orgLoading, refresh } = useOrg();
  const [step, setStep] = useState(1);
  const [workspaceName, setWorkspaceName] = useState("");
  const [saving, setSaving] = useState(false);
  const [initialLoaded, setInitialLoaded] = useState(false);

  // Wait for org context to load after signup
  useEffect(() => {
    if (!orgLoading && !currentOrg && !initialLoaded) {
      // Org context not loaded yet, retry
      const timer = setTimeout(() => refresh(), 500);
      return () => clearTimeout(timer);
    }
    if (currentOrg) {
      setInitialLoaded(true);
    }
  }, [orgLoading, currentOrg, initialLoaded, refresh]);

  async function handleSetupWorkspace(e: React.FormEvent) {
    e.preventDefault();
    if (!workspaceName.trim() || !currentOrg) return;
    setSaving(true);
    try {
      await orgClient.updateOrganization({
        id: currentOrg.id,
        name: workspaceName.trim(),
      });
      await refresh();
      setStep(2);
    } catch {
      // handle error
    } finally {
      setSaving(false);
    }
  }

  function handleFinish() {
    router.push("/projects");
  }

  return (
    <div className="flex min-h-screen items-center justify-center px-4">
      <div className="w-full max-w-lg">
        {/* Progress */}
        <div className="mb-8 flex items-center justify-center gap-2">
          {[1, 2].map((s) => (
            <div
              key={s}
              className={`h-1.5 w-12 rounded-full transition-colors ${
                s <= step ? "bg-primary" : "bg-border"
              }`}
            />
          ))}
        </div>

        {step === 1 && (
          <div className="space-y-6">
            <div className="text-center">
              <h1 className="text-2xl font-bold tracking-tight">Create your workspace</h1>
              <p className="mt-2 text-sm text-muted-foreground">
                A workspace is where your team collaborates on specs and designs.
              </p>
            </div>

            <form onSubmit={handleSetupWorkspace} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="workspace-name">Workspace name</Label>
                <Input
                  id="workspace-name"
                  value={workspaceName}
                  onChange={(e) => setWorkspaceName(e.target.value)}
                  placeholder="Acme Inc."
                  required
                  autoFocus
                  className="h-11 text-base"
                />
                <p className="text-xs text-muted-foreground">
                  You can change this later in settings.
                </p>
              </div>

              <Button
                type="submit"
                className="w-full cursor-pointer h-11"
                disabled={saving || !workspaceName.trim()}
              >
                {saving ? "Setting up..." : "Continue"}
              </Button>
            </form>
          </div>
        )}

        {step === 2 && (
          <div className="space-y-6 text-center">
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-primary/10">
              <svg
                className="h-8 w-8 text-primary"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </div>

            <div>
              <h1 className="text-2xl font-bold tracking-tight">You&apos;re all set!</h1>
              <p className="mt-2 text-sm text-muted-foreground">
                Your workspace <span className="font-medium text-foreground">{workspaceName}</span>{" "}
                is ready. Start by creating your first project.
              </p>
            </div>

            <Button onClick={handleFinish} className="cursor-pointer h-11 px-8">
              Go to projects
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}
