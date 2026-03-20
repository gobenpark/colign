"use client";

import { useState } from "react";
import { Circle, Loader, CheckCircle, User, Trash2 } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

const statusIcons = {
  todo: Circle,
  in_progress: Loader,
  done: CheckCircle,
} as const;

const statusColors = {
  todo: "text-muted-foreground",
  in_progress: "text-yellow-500",
  done: "text-green-500",
} as const;

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
  const [confirmDelete, setConfirmDelete] = useState(false);

  // Local edit state — initialized from task prop
  const [title, setTitle] = useState(task.title);
  const [description, setDescription] = useState(task.description);
  const [specRef, setSpecRef] = useState(task.specRef);

  const status = task.status as keyof typeof statusIcons;
  const StatusIcon = statusIcons[status] ?? Circle;
  const statusColor = statusColors[status] ?? statusColors.todo;

  function handleStatusCycle(e: React.MouseEvent) {
    e.stopPropagation();
    onUpdate(task.id, { status: nextStatus[task.status] ?? "todo" });
  }

  function handleRowClick() {
    setExpanded((prev) => !prev);
    setConfirmDelete(false);
  }

  function handleTitleBlur() {
    if (title.trim() !== task.title) {
      onUpdate(task.id, { title: title.trim() });
    }
  }

  function handleDescriptionBlur() {
    if (description !== task.description) {
      onUpdate(task.id, { description });
    }
  }

  function handleSpecRefBlur() {
    if (specRef !== task.specRef) {
      onUpdate(task.id, { specRef });
    }
  }

  function handleStatusChange(value: string) {
    onUpdate(task.id, { status: value });
  }

  function handleAssigneeChange(value: string) {
    if (value === "unassigned") {
      onUpdate(task.id, { assigneeId: null });
    } else {
      onUpdate(task.id, { assigneeId: BigInt(value) });
    }
  }

  function handleDeleteClick(e: React.MouseEvent) {
    e.stopPropagation();
    if (confirmDelete) {
      onDelete(task.id);
    } else {
      setConfirmDelete(true);
    }
  }

  const isDone = task.status === "done";

  return (
    <div
      className={cn(
        "border border-border/50 rounded-md bg-card transition-colors duration-200",
        isDone && "opacity-70",
      )}
    >
      {/* Collapsed row */}
      <div
        className="flex items-center gap-2 px-2 py-1 cursor-pointer hover:bg-muted/30 transition-colors duration-200 rounded-md"
        onClick={handleRowClick}
      >
        {/* Status icon — 44x44 touch target via padding */}
        <button
          type="button"
          onClick={handleStatusCycle}
          className="flex items-center justify-center p-2.5 rounded-md hover:bg-muted/50 transition-colors duration-200 shrink-0"
          style={{ minWidth: 44, minHeight: 44 }}
          aria-label={t("tasks.statusTodo")}
        >
          <StatusIcon className={cn("h-4 w-4", statusColor)} />
        </button>

        {/* Title */}
        <span
          className={cn(
            "text-sm font-medium flex-1 min-w-0 truncate",
            isDone && "line-through text-muted-foreground",
          )}
        >
          {task.title || t("tasks.titlePlaceholder")}
        </span>

        {/* Description (truncated) */}
        {task.description && (
          <span className="text-xs text-muted-foreground truncate max-w-[200px] hidden sm:block">
            {task.description}
          </span>
        )}

        {/* Spec ref badge */}
        {task.specRef && (
          <span className="text-xs font-mono bg-muted text-muted-foreground px-1.5 py-0.5 rounded shrink-0">
            {task.specRef}
          </span>
        )}

        {/* Assignee initials */}
        {task.assigneeName ? (
          <span className="flex items-center justify-center h-6 w-6 rounded-full bg-primary/20 text-primary text-xs font-semibold shrink-0">
            {getInitials(task.assigneeName)}
          </span>
        ) : (
          <span className="flex items-center justify-center h-6 w-6 rounded-full bg-muted shrink-0">
            <User className="h-3 w-3 text-muted-foreground" />
          </span>
        )}
      </div>

      {/* Expanded inline edit form */}
      {expanded && (
        <div
          className="px-4 pb-4 pt-2 border-t border-border/30 flex flex-col gap-3"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Title */}
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onBlur={handleTitleBlur}
            placeholder={t("tasks.titlePlaceholder")}
            className="w-full rounded-md border border-border bg-transparent px-3 py-2 text-sm outline-none focus:border-primary transition-colors duration-200"
          />

          {/* Description */}
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            onBlur={handleDescriptionBlur}
            placeholder={t("tasks.descriptionPlaceholder")}
            rows={2}
            className="w-full rounded-md border border-border bg-transparent px-3 py-2 text-sm outline-none focus:border-primary transition-colors duration-200 resize-none"
          />

          <div className="flex flex-wrap gap-2">
            {/* Status dropdown */}
            <Select value={task.status} onValueChange={handleStatusChange}>
              <SelectTrigger size="sm" className="w-36">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="todo">{t("tasks.statusTodo")}</SelectItem>
                <SelectItem value="in_progress">{t("tasks.statusInProgress")}</SelectItem>
                <SelectItem value="done">{t("tasks.statusDone")}</SelectItem>
              </SelectContent>
            </Select>

            {/* Spec ref */}
            <input
              type="text"
              value={specRef}
              onChange={(e) => setSpecRef(e.target.value)}
              onBlur={handleSpecRefBlur}
              placeholder={t("tasks.specRefPlaceholder")}
              className="w-28 rounded-md border border-border bg-transparent px-2 py-1 text-sm font-mono outline-none focus:border-primary transition-colors duration-200"
            />

            {/* Assignee dropdown */}
            <Select
              value={task.assigneeId != null ? String(task.assigneeId) : "unassigned"}
              onValueChange={handleAssigneeChange}
            >
              <SelectTrigger size="sm" className="w-40">
                <SelectValue placeholder={t("tasks.assignee")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="unassigned">{t("tasks.unassigned")}</SelectItem>
                {members.map((m) => (
                  <SelectItem key={String(m.userId)} value={String(m.userId)}>
                    {m.userName}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>

            {/* Delete button */}
            <button
              type="button"
              onClick={handleDeleteClick}
              className={cn(
                "ml-auto flex items-center gap-1.5 rounded-md px-2 py-1 text-xs transition-colors duration-200",
                confirmDelete
                  ? "bg-destructive/20 text-destructive hover:bg-destructive/30"
                  : "text-muted-foreground hover:text-destructive hover:bg-destructive/10",
              )}
            >
              <Trash2 className="h-3.5 w-3.5" />
              {confirmDelete ? t("tasks.deleteConfirm") : t("common.delete")}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
