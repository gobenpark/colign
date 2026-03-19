"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Input } from "@/components/ui/input";
import { useI18n } from "@/lib/i18n";

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

interface WorkflowPanelProps {
  stage: string;
  conditions: GateCondition[];
  history: WorkflowEvent[];
  onApprove: () => void;
  onRequestChanges: () => void;
  onRevert: (reason: string) => void;
}

const stageConfig: Record<string, { label: string; color: string }> = {
  draft: { label: "Draft", color: "text-yellow-400" },
  design: { label: "Design", color: "text-blue-400" },
  review: { label: "Review", color: "text-purple-400" },
  ready: { label: "Ready", color: "text-emerald-400" },
};

export function WorkflowPanel({ stage, conditions, history, onApprove, onRequestChanges, onRevert }: WorkflowPanelProps) {
  const { t } = useI18n();
  const [revertReason, setRevertReason] = useState("");
  const [showRevert, setShowRevert] = useState(false);
  const currentConfig = stageConfig[stage] ?? stageConfig.draft;

  return (
    <div className="grid gap-6 lg:grid-cols-3">
      <div className="lg:col-span-2">
        <Card className="border-border/50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <span className={currentConfig.color}>{t(`stages.${stage}`)}</span>
              {t("change.gateConditions")}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-3">
              {conditions.map((c) => (
                <li key={c.name} className="flex items-center gap-3 rounded-lg border border-border/50 p-3">
                  <div className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full ${c.met ? "bg-emerald-500/10" : "bg-destructive/10"}`}>
                    {c.met ? (
                      <svg className="h-3.5 w-3.5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                      </svg>
                    ) : (
                      <svg className="h-3.5 w-3.5 text-destructive" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    )}
                  </div>
                  <span className="text-sm">{c.description}</span>
                </li>
              ))}
            </ul>
            <Separator className="my-5" />
            <div className="flex flex-wrap gap-2">
              {stage === "review" && (
                <>
                  <Button onClick={onApprove} size="sm" className="cursor-pointer">{t("change.approve")}</Button>
                  <Button onClick={onRequestChanges} variant="outline" size="sm" className="cursor-pointer">{t("change.requestChanges")}</Button>
                </>
              )}
              {stage !== "draft" && (
                <Button variant="ghost" size="sm" className="cursor-pointer text-muted-foreground" onClick={() => setShowRevert(!showRevert)}>{t("change.revert")}</Button>
              )}
            </div>
            {showRevert && (
              <div className="mt-3 flex gap-2">
                <Input placeholder={t("change.revertReason")} value={revertReason} onChange={(e) => setRevertReason(e.target.value)} className="flex-1" />
                <Button onClick={() => { onRevert(revertReason); setShowRevert(false); setRevertReason(""); }} size="sm" disabled={!revertReason.trim()} className="cursor-pointer">{t("common.confirm")}</Button>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
      <div>
        <Card className="border-border/50">
          <CardHeader><CardTitle className="text-base">{t("change.history")}</CardTitle></CardHeader>
          <CardContent>
            {history.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("change.noEvents")}</p>
            ) : (
              <ul className="space-y-4">
                {history.map((event) => (
                  <li key={String(event.id)} className="relative pl-5">
                    <div className="absolute left-0 top-1.5 h-2 w-2 rounded-full bg-primary/50" />
                    <p className="text-sm font-medium">{event.action.replace("_", " ")}</p>
                    <p className="text-xs text-muted-foreground">{event.fromStage} → {event.toStage}</p>
                    {event.reason && <p className="mt-0.5 text-xs text-muted-foreground">{event.reason}</p>}
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
