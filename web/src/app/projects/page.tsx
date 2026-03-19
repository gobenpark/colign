"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Header } from "@/components/layout/header";
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
          }))
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
      <Header
        actions={
          <Link href="/projects/new">
            <Button size="sm" className="cursor-pointer">
              <svg className="mr-1.5 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              {t("projects.newProject")}
            </Button>
          </Link>
        }
      />

      <main className="mx-auto max-w-6xl px-6 py-10">
        <h1 className="mb-8 text-2xl font-semibold tracking-tight">{t("projects.title")}</h1>

        {projects.length === 0 ? (
          <Card className="border-dashed border-border/50">
            <CardHeader className="items-center py-16 text-center">
              <div className="mb-4 rounded-full bg-muted p-4">
                <svg className="h-8 w-8 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
                </svg>
              </div>
              <CardTitle className="text-lg">{t("projects.noProjects")}</CardTitle>
              <CardDescription className="mt-1">
                {t("projects.createFirst")}
              </CardDescription>
              <Link href="/projects/new" className="mt-6">
                <Button className="cursor-pointer">{t("projects.createProject")}</Button>
              </Link>
            </CardHeader>
          </Card>
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
