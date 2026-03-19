"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { useI18n } from "@/lib/i18n";
import { workflowClient } from "@/lib/workflow";
import { WorkflowPanel } from "@/components/change/workflow-panel";
import { DocumentTab } from "@/components/change/document-tab";
import { ChatTab } from "@/components/change/chat-tab";

interface GateCondition {
  name: string;
  description: string;
  met: boolean;
}

interface WorkflowEvent {
  id: bigint;
  fromStage: string;
  toStage: string;
  action: string;
  reason: string;
}

const stageConfig: Record<string, { label: string; color: string; icon: string }> = {
  draft: { label: "Draft", color: "text-yellow-400", icon: "M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" },
  design: { label: "Design", color: "text-blue-400", icon: "M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zm0 8a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6z" },
  review: { label: "Review", color: "text-purple-400", icon: "M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" },
  ready: { label: "Ready", color: "text-emerald-400", icon: "M5 13l4 4L19 7" },
};

const stages = ["draft", "design", "review", "ready"];

type TabId = "workflow" | "proposal" | "design" | "specs" | "tasks";

const tabI18nKeys: Record<TabId, string> = {
  workflow: "change.workflow",
  proposal: "change.proposal",
  design: "change.design",
  specs: "change.specs",
  tasks: "change.tasks",
};

export default function ChangeDetailPage() {
  const params = useParams();
  const slug = params.slug as string;
  const changeId = BigInt(params.changeId as string);
  const { t } = useI18n();

  const [activeTab, setActiveTab] = useState<TabId>("workflow");
  const [chatOpen, setChatOpen] = useState(false);
  const [stage, setStage] = useState("");
  const [conditions, setConditions] = useState<GateCondition[]>([]);
  const [history, setHistory] = useState<WorkflowEvent[]>([]);
  const [loading, setLoading] = useState(true);

  async function loadAll() {
    try {
      const [statusRes, historyRes] = await Promise.all([
        workflowClient.getStatus({ changeId }),
        workflowClient.getHistory({ changeId }),
      ]);
      setStage(statusRes.stage);
      setConditions(statusRes.conditions.map((c) => ({ name: c.name, description: c.description, met: c.met })));
      setHistory(historyRes.events.map((e) => ({
        id: e.id, fromStage: e.fromStage, toStage: e.toStage, action: e.action, reason: e.reason,
      })));
    } catch {
      // handle error
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => { loadAll(); }, []);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  const currentIdx = stages.indexOf(stage);

  return (
    <div className="flex min-h-screen flex-col">
      {/* Header */}
      <header className="sticky top-0 z-30 border-b border-border/50 bg-background/80 backdrop-blur-md">
        <div className="flex items-center justify-between px-6 py-3">
          <div className="flex items-center gap-3">
            <Link href="/projects" className="text-xl font-bold tracking-tight">
              Co<span className="text-primary">Spec</span>
            </Link>
            <svg className="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.25 4.5l7.5 7.5-7.5 7.5" />
            </svg>
            <Link href={`/projects/${slug}`} className="text-sm text-muted-foreground hover:text-foreground transition-colors duration-200">
              Project
            </Link>
            <svg className="h-4 w-4 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.25 4.5l7.5 7.5-7.5 7.5" />
            </svg>
            <span className="text-sm font-medium">Change</span>
          </div>

          {/* Chat Toggle */}
          <Button
            variant={chatOpen ? "default" : "outline"}
            size="sm"
            onClick={() => setChatOpen(!chatOpen)}
            className="cursor-pointer gap-1.5"
          >
            <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8.625 12a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375M21 12c0 4.556-4.03 8.25-9 8.25a9.764 9.764 0 01-2.555-.337A5.972 5.972 0 015.41 20.97a5.969 5.969 0 01-.474-.065 4.48 4.48 0 00.978-2.025c.09-.457-.133-.901-.467-1.226C3.93 16.178 3 14.189 3 12c0-4.556 4.03-8.25 9-8.25s9 3.694 9 8.25z" />
            </svg>
            {t("change.aiChat")}
          </Button>
        </div>
      </header>

      {/* Main Layout */}
      <div className="flex flex-1 overflow-hidden">
        {/* Left: Content */}
        <div className={`flex-1 overflow-y-auto transition-all duration-300 ${chatOpen ? "mr-0" : ""}`}>
          <div className="mx-auto max-w-5xl px-6 py-6">
            {/* Stage Progress */}
            <div className="mb-8">
              <div className="flex items-center justify-between">
                {stages.map((s, i) => {
                  const cfg = stageConfig[s];
                  const isActive = i === currentIdx;
                  const isPast = i < currentIdx;
                  return (
                    <div key={s} className="flex flex-1 items-center">
                      <div className="flex flex-col items-center">
                        <div className={`flex h-9 w-9 items-center justify-center rounded-full border-2 transition-all duration-300 ${
                          isActive ? "border-primary bg-primary/10" : isPast ? "border-emerald-500 bg-emerald-500/10" : "border-border bg-muted"
                        }`}>
                          <svg className={`h-4 w-4 ${isActive ? cfg.color : isPast ? "text-emerald-400" : "text-muted-foreground"}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={isPast ? "M5 13l4 4L19 7" : cfg.icon} />
                          </svg>
                        </div>
                        <span className={`mt-1.5 text-[11px] font-medium ${isActive ? "text-foreground" : "text-muted-foreground"}`}>
                          {t(`stages.${s}`)}
                        </span>
                      </div>
                      {i < stages.length - 1 && (
                        <div className={`mx-2 h-0.5 flex-1 rounded-full transition-colors duration-300 ${isPast ? "bg-emerald-500/50" : "bg-border"}`} />
                      )}
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Tab Navigation */}
            <div className="mb-6 flex gap-1 overflow-x-auto border-b border-border/50">
              {(Object.keys(tabI18nKeys) as TabId[]).map((tabId) => (
                <button
                  key={tabId}
                  onClick={() => setActiveTab(tabId)}
                  className={`cursor-pointer whitespace-nowrap px-4 py-2.5 text-sm font-medium transition-colors duration-200 ${
                    activeTab === tabId
                      ? "border-b-2 border-primary text-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  }`}
                >
                  {t(tabI18nKeys[tabId])}
                </button>
              ))}
            </div>

            {/* Tab Content */}
            {activeTab === "workflow" && (
              <WorkflowPanel
                stage={stage}
                conditions={conditions}
                history={history}
                onApprove={async () => {
                  await workflowClient.approve({ changeId, comment: "" });
                  loadAll();
                }}
                onRequestChanges={async () => {
                  await workflowClient.requestChanges({ changeId, reason: "Changes needed" });
                  loadAll();
                }}
                onRevert={async (reason) => {
                  await workflowClient.revert({ changeId, reason });
                  loadAll();
                }}
              />
            )}

            {activeTab === "proposal" && <DocumentTab changeId={changeId} docType="proposal" />}
            {activeTab === "design" && <DocumentTab changeId={changeId} docType="design" />}
            {activeTab === "specs" && <DocumentTab changeId={changeId} docType="spec" />}
            {activeTab === "tasks" && <DocumentTab changeId={changeId} docType="tasks" />}
          </div>
        </div>

        {/* Right: Chat Side Panel */}
        <div
          className={`border-l border-border/50 bg-card/30 transition-all duration-300 ${
            chatOpen ? "w-[400px] min-w-[400px]" : "w-0 min-w-0 overflow-hidden border-l-0"
          }`}
        >
          {chatOpen && (
            <div className="flex h-full flex-col">
              <div className="flex items-center justify-between border-b border-border/50 px-4 py-3">
                <div className="flex items-center gap-2">
                  <h3 className="text-sm font-medium">{t("change.aiChat")}</h3>
                  <span className="inline-flex h-4 items-center rounded-full bg-primary/10 px-1.5 text-[10px] font-medium text-primary">
                    AI
                  </span>
                </div>
                <button
                  onClick={() => setChatOpen(false)}
                  className="cursor-pointer rounded p-1 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                >
                  <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
              <div className="flex-1 overflow-hidden px-4">
                <ChatTab changeId={changeId} />
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
