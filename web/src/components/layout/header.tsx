"use client";

import { useRef, useState, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import { useOrg } from "@/lib/org-context";

interface Breadcrumb {
  label: string;
  href?: string;
}

interface HeaderProps {
  breadcrumbs?: Breadcrumb[];
  actions?: React.ReactNode;
}

export function Header({ breadcrumbs = [], actions }: HeaderProps) {
  const pathname = usePathname();
  const isSettings = pathname.startsWith("/settings");
  const { currentOrg, orgs, switchOrg } = useOrg();
  const [orgMenuOpen, setOrgMenuOpen] = useState(false);
  const orgMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (orgMenuRef.current && !orgMenuRef.current.contains(e.target as Node)) {
        setOrgMenuOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  return (
    <header className="sticky top-0 z-30 border-b border-border/50 bg-background/80 backdrop-blur-md">
      <div className="flex items-center justify-between px-6 py-3">
        <div className="flex items-center gap-3">
          <Link href="/projects" className="text-xl font-bold tracking-tight">
            Co<span className="text-primary">lign</span>
          </Link>

          {/* Org Switcher */}
          {currentOrg && (
            <>
              <svg
                className="h-3.5 w-3.5 text-muted-foreground/40"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M8.25 4.5l7.5 7.5-7.5 7.5"
                />
              </svg>
              <div className="relative flex items-center gap-1" ref={orgMenuRef}>
                <Link
                  href="/projects"
                  className="flex items-center gap-1.5 rounded-lg px-2 py-1 text-sm font-medium text-foreground/80 transition-colors hover:bg-accent hover:text-foreground"
                >
                  <svg
                    className="h-3.5 w-3.5 text-muted-foreground"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={1.5}
                      d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21"
                    />
                  </svg>
                  {currentOrg.name}
                </Link>
                {orgs.length > 1 && (
                  <button
                    onClick={() => setOrgMenuOpen(!orgMenuOpen)}
                    className="cursor-pointer rounded p-0.5 text-muted-foreground/50 transition-colors hover:bg-accent hover:text-foreground"
                  >
                    <svg
                      className="h-3 w-3"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19.5 8.25l-7.5 7.5-7.5-7.5"
                      />
                    </svg>
                  </button>
                )}
                {orgMenuOpen && orgs.length > 1 && (
                  <div className="absolute left-0 top-9 z-50 w-56 rounded-xl border border-border/50 bg-popover p-1.5 shadow-xl animate-in fade-in slide-in-from-top-2 duration-150">
                    {orgs.map((org) => (
                      <button
                        key={String(org.id)}
                        onClick={() => {
                          setOrgMenuOpen(false);
                          if (org.id !== currentOrg.id) {
                            switchOrg(org.id);
                          }
                        }}
                        className={`flex w-full cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 text-sm transition-colors ${
                          org.id === currentOrg.id
                            ? "bg-accent text-foreground"
                            : "text-foreground/80 hover:bg-accent"
                        }`}
                      >
                        <svg
                          className="h-3.5 w-3.5 text-muted-foreground"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={1.5}
                            d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21"
                          />
                        </svg>
                        <span className="truncate">{org.name}</span>
                        {org.id === currentOrg.id && (
                          <svg
                            className="ml-auto h-3.5 w-3.5 text-primary"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M4.5 12.75l6 6 9-13.5"
                            />
                          </svg>
                        )}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            </>
          )}

          {breadcrumbs.map((crumb, i) => (
            <div key={i} className="flex items-center gap-3">
              <svg
                className="h-3.5 w-3.5 text-muted-foreground/40"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M8.25 4.5l7.5 7.5-7.5 7.5"
                />
              </svg>
              {crumb.href ? (
                <Link
                  href={crumb.href}
                  className="text-sm text-muted-foreground transition-colors duration-200 hover:text-foreground"
                >
                  {crumb.label}
                </Link>
              ) : (
                <span className="text-sm font-medium">{crumb.label}</span>
              )}
            </div>
          ))}
        </div>

        <div className="flex items-center gap-2">
          {actions}
          <Link href="/settings">
            <Button
              variant="ghost"
              size="sm"
              className={`cursor-pointer ${isSettings ? "text-foreground" : "text-muted-foreground"}`}
            >
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z"
                />
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                />
              </svg>
            </Button>
          </Link>
        </div>
      </div>
    </header>
  );
}
