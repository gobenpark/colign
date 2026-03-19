"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { projectClient } from "@/lib/project";

const stageBadge: Record<string, { label: string; color: string }> = {
  draft: { label: "Draft", color: "bg-yellow-500/10 text-yellow-400 border border-yellow-500/20" },
  design: { label: "Design", color: "bg-blue-500/10 text-blue-400 border border-blue-500/20" },
  review: { label: "Review", color: "bg-purple-500/10 text-purple-400 border border-purple-500/20" },
  ready: { label: "Ready", color: "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20" },
};

interface Change {
  id: bigint;
  name: string;
  stage: string;
}

interface ProjectDetail {
  id: bigint;
  name: string;
  slug: string;
  description: string;
}

export default function ProjectDetailPage() {
  const params = useParams();
  const slug = params.slug as string;

  const [project, setProject] = useState<ProjectDetail | null>(null);
  const [changes, setChanges] = useState<Change[]>([]);
  const [newChangeName, setNewChangeName] = useState("");
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    async function load() {
      try {
        const projectRes = await projectClient.getProject({ slug });
        if (projectRes.project) {
          setProject({
            id: projectRes.project.id,
            name: projectRes.project.name,
            slug: projectRes.project.slug,
            description: projectRes.project.description,
          });

          const changesRes = await projectClient.listChanges({
            projectId: projectRes.project.id,
          });
          setChanges(
            changesRes.changes.map((c) => ({
              id: c.id,
              name: c.name,
              stage: c.stage,
            }))
          );
        }
      } catch {
        // handle error
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [slug]);

  async function handleCreateChange(e: React.FormEvent) {
    e.preventDefault();
    if (!newChangeName.trim() || !project) return;
    setCreating(true);
    try {
      const res = await projectClient.createChange({
        projectId: project.id,
        name: newChangeName.trim(),
      });
      if (res.change) {
        setChanges((prev) => [
          { id: res.change!.id, name: res.change!.name, stage: res.change!.stage },
          ...prev,
        ]);
        setNewChangeName("");
      }
    } catch {
      // handle error
    } finally {
      setCreating(false);
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  if (!project) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p className="text-muted-foreground">Project not found</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      {/* Top Bar */}
      <header className="sticky top-0 z-30 border-b border-border/50 bg-background/80 backdrop-blur-md">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-3">
            <Link href="/projects" className="text-xl font-bold tracking-tight">
              Co<span className="text-primary">lign</span>
            </Link>
            <svg className="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.25 4.5l7.5 7.5-7.5 7.5" />
            </svg>
            <span className="text-sm font-medium">{project.name}</span>
          </div>
          <Link href={`/projects/${project.slug}/settings`}>
            <Button variant="ghost" size="sm" className="cursor-pointer text-muted-foreground">
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </Button>
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-10">
        {/* Project Info */}
        {project.description && (
          <p className="mb-8 text-muted-foreground">{project.description}</p>
        )}

        {/* Changes Section */}
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Changes</h2>
        </div>

        <form onSubmit={handleCreateChange} className="mb-6 flex gap-2">
          <Input
            value={newChangeName}
            onChange={(e) => setNewChangeName(e.target.value)}
            placeholder="New change name (e.g., add-user-auth)"
            className="flex-1"
          />
          <Button type="submit" disabled={creating || !newChangeName.trim()} className="cursor-pointer">
            {creating ? "Creating..." : "Create"}
          </Button>
        </form>

        {changes.length === 0 ? (
          <Card className="border-dashed border-border/50">
            <CardContent className="py-12 text-center">
              <p className="text-muted-foreground">No changes yet. Create one to start the SDD workflow.</p>
            </CardContent>
          </Card>
        ) : (
          <div className="space-y-3">
            {changes.map((change) => {
              const badge = stageBadge[change.stage] ?? stageBadge.draft;
              return (
                <Link key={String(change.id)} href={`/projects/${project.slug}/changes/${change.id}`}>
                  <Card className="cursor-pointer border-border/50 transition-all duration-200 hover:border-primary/30 hover:bg-card/80">
                    <CardHeader className="flex flex-row items-center justify-between py-4">
                      <div>
                        <CardTitle className="text-base font-medium">{change.name}</CardTitle>
                        <CardDescription className="mt-1.5">
                          <span className={`inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium ${badge.color}`}>
                            {badge.label}
                          </span>
                        </CardDescription>
                      </div>
                      <svg className="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.25 4.5l7.5 7.5-7.5 7.5" />
                      </svg>
                    </CardHeader>
                  </Card>
                </Link>
              );
            })}
          </div>
        )}
      </main>
    </div>
  );
}
