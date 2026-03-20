"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useI18n } from "@/lib/i18n";
import { projectClient } from "@/lib/project";

interface Project {
  id: bigint;
  name: string;
  slug: string;
  description: string;
}

export default function ProjectsPage() {
  const { t } = useI18n();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetchProjects() {
      try {
        const res = await projectClient.listProjects({});
        setProjects(
          res.projects.map((p) => ({
            id: p.id,
            name: p.name,
            slug: p.slug,
            description: p.description,
          })),
        );
      } catch {
        // handle error
      } finally {
        setLoading(false);
      }
    }
    fetchProjects();
  }, []);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <main className="mx-auto max-w-6xl px-6 py-10">
        <div className="mb-8 flex items-center justify-between">
          <h1 className="text-2xl font-semibold tracking-tight">{t("projects.title")}</h1>
          <Link href="/projects/new">
            <Button size="sm" className="cursor-pointer">
              <svg className="mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              {t("projects.newProject")}
            </Button>
          </Link>
        </div>

        {projects.length === 0 ? (
          <div className="flex flex-col items-center justify-center rounded-2xl border border-dashed border-border/40 bg-card/30 py-20">
            <div className="rounded-2xl bg-primary/5 p-5 mb-5">
              <svg
                className="h-10 w-10 text-primary/40"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1}
                  d="M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z"
                />
              </svg>
            </div>
            <p className="text-sm font-medium text-foreground/70">{t("projects.noProjects")}</p>
            <p className="mt-1 text-xs text-muted-foreground">{t("projects.createFirst")}</p>
            <Link href="/projects/new" className="mt-6">
              <Button className="cursor-pointer">{t("projects.createProject")}</Button>
            </Link>
          </div>
        ) : (
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {projects.map((project) => (
              <Link key={String(project.id)} href={`/projects/${project.slug}`}>
                <Card className="cursor-pointer border-border/50 transition-all duration-200 hover:border-primary/30 hover:bg-card/80">
                  <CardHeader>
                    <CardTitle className="text-base">{project.name}</CardTitle>
                    {project.description && (
                      <CardDescription className="line-clamp-2 text-sm">
                        {project.description}
                      </CardDescription>
                    )}
                  </CardHeader>
                </Card>
              </Link>
            ))}
          </div>
        )}
      </main>
    </div>
  );
}
