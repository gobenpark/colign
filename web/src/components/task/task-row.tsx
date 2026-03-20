"use client";

import { useState, useRef, useEffect } from "react";
import { Circle, Clock, CheckCircle2, User, Trash2, X } from "lucide-react";
import { useI18n } from "@/lib/i18n";

const statusConfig: Record<string, { icon: typeof Circle; color: string; bg: string }> = {
  todo: { icon: Circle, color: "text-muted-foreground", bg: "bg-muted/50" },
  in_progress: { icon: Clock, color: "text-blue-400", bg: "bg-blue-400/10" },
  done: { icon: CheckCircle2, color: "text-emerald-400", bg: "bg-emerald-400/10" },
};

const nextStatus: Record<string, string> = {
  todo: "in_progress",
  in_progress: "done",
  done: "todo",
};

interface TaskRowProps {
  task: {
    id: bigint;
    title: string;
    description: string;
    status: string;
    specRef: string;
    assigneeId?: bigint;
    assigneeName: string;
    orderIndex: number;
  };
  members: Array<{ userId: bigint; userName: string }>;
  onUpdate: (id: bigint, fields: Record<string, unknown>) => void;
  onDelete: (id: bigint) => void;
}

function getInitials(name: string): string {
  if (!name) return "";
  return name
    .split(" ")
    .map((part) => part[0])
    .join("")
    .toUpperCase()
    .slice(0, 2);
}

export function TaskRow({ task, members, onUpdate, onDelete }: TaskRowProps) {
  const { t } = useI18n();
  const [expanded, setExpanded] = useState(false);
  const [title, setTitle] = useState(task.title);
  const [description, setDescription] = useState(task.description);
  const rowRef = useRef<HTMLDivElement>(null);

  const cfg = statusConfig[task.status] || statusConfig.todo;
  const StatusIcon = cfg.icon;
  const isDone = task.status === "done";

  // Close on outside click
  useEffect(() => {
    if (!expanded) return;
    function handleClickOutside(e: MouseEvent) {
      if (rowRef.current && !rowRef.current.contains(e.target as Node)) {
        setExpanded(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [expanded]);

  function cycleStatus(e: React.MouseEvent) {
    e.stopPropagation();
    onUpdate(task.id, { status: nextStatus[task.status] ?? "todo" });
  }

  function handleTitleBlur() {
    if (title.trim() !== task.title) onUpdate(task.id, { title: title.trim() });
  }

  function handleDescBlur() {
    if (description !== task.description) onUpdate(task.id, { description });
  }

  if (expanded) {
    return (
      <div ref={rowRef} className="rounded-lg border border-primary/20 bg-card/80 shadow-md">
        {/* Header */}
        <div className="flex items-center gap-2.5 border-b border-border/30 px-4 py-2.5">
          <button
            onClick={cycleStatus}
            className={`cursor-pointer rounded-md p-1 ${cfg.bg} transition-colors hover:opacity-80`}
          >
            <StatusIcon className={`size-3.5 ${cfg.color}`} />
          </button>
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onBlur={handleTitleBlur}
            placeholder={t("tasks.titlePlaceholder")}
            className="flex-1 bg-transparent text-sm font-medium outline-none placeholder:text-muted-foreground/40"
            autoFocus
          />
          <button
            onClick={() => setExpanded(false)}
            className="cursor-pointer rounded-md p-1 text-muted-foreground/50 hover:bg-accent hover:text-foreground transition-colors"
          >
            <X className="size-3.5" />
          </button>
        </div>

        {/* Body */}
        <div className="px-4 py-3 space-y-3">
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            onBlur={handleDescBlur}
            placeholder={t("tasks.descriptionPlaceholder")}
            rows={2}
            className="w-full resize-none bg-transparent text-sm text-foreground/70 outline-none placeholder:text-muted-foreground/40"
          />

          {/* Properties */}
          <div className="flex flex-wrap items-center gap-2 pt-1">
            <select
              value={task.assigneeId !== undefined ? String(task.assigneeId) : ""}
              onChange={(e) => {
                const v = e.target.value;
                onUpdate(task.id, { assigneeId: v === "" ? null : BigInt(v) });
              }}
              className="cursor-pointer rounded-md border border-border/40 bg-transparent px-2 py-1 text-xs text-foreground/70 outline-none hover:border-border/60 transition-colors"
            >
              <option value="">{t("tasks.unassigned")}</option>
              {members.map((m) => (
                <option key={String(m.userId)} value={String(m.userId)}>
                  {m.userName}
                </option>
              ))}
            </select>

            <input
              value={task.specRef}
              onChange={(e) => onUpdate(task.id, { specRef: e.target.value })}
              placeholder={t("tasks.specRefPlaceholder")}
              className="w-20 rounded-md border border-border/40 bg-transparent px-2 py-1 text-xs text-foreground/70 outline-none placeholder:text-muted-foreground/30 hover:border-border/60 focus:border-primary/50 transition-colors"
            />

            <button
              onClick={() => onDelete(task.id)}
              className="ml-auto cursor-pointer rounded-md p-1 text-muted-foreground/40 hover:bg-destructive/10 hover:text-destructive transition-colors"
            >
              <Trash2 className="size-3.5" />
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Collapsed — single row
  return (
    <div
      ref={rowRef}
      onClick={() => setExpanded(true)}
      className={`group flex cursor-pointer items-center gap-2.5 rounded-lg px-3 py-2 transition-colors hover:bg-accent/30 ${
        isDone ? "opacity-60" : ""
      }`}
    >
      <button
        onClick={cycleStatus}
        className={`shrink-0 cursor-pointer rounded-md p-0.5 ${cfg.bg} transition-colors hover:opacity-80`}
      >
        <StatusIcon className={`size-3.5 ${cfg.color}`} />
      </button>

      <span
        className={`min-w-0 flex-1 truncate text-sm ${
          isDone ? "line-through text-muted-foreground" : "text-foreground/90"
        }`}
      >
        {task.title}
      </span>

      {task.specRef && (
        <span className="shrink-0 rounded bg-muted/60 px-1.5 py-0.5 text-[10px] text-muted-foreground">
          {task.specRef}
        </span>
      )}

      {task.assigneeName ? (
        <div className="flex size-6 shrink-0 items-center justify-center rounded-full bg-primary/10 text-[10px] font-medium text-primary">
          {getInitials(task.assigneeName)}
        </div>
      ) : (
        <div className="flex size-6 shrink-0 items-center justify-center rounded-full border border-dashed border-border/40 text-muted-foreground/30 opacity-0 group-hover:opacity-100 transition-opacity">
          <User className="size-3" />
        </div>
      )}
    </div>
  );
}
