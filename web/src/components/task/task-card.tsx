"use client";

import { useState, useRef, useEffect } from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical, User, Trash2, X, Circle, Clock, CheckCircle2 } from "lucide-react";
import { useI18n } from "@/lib/i18n";

interface TaskCardProps {
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
  isDragging?: boolean;
}

const statusConfig: Record<string, { icon: typeof Circle; color: string; bg: string }> = {
  todo: { icon: Circle, color: "text-muted-foreground", bg: "bg-muted/50" },
  in_progress: { icon: Clock, color: "text-blue-400", bg: "bg-blue-400/10" },
  done: { icon: CheckCircle2, color: "text-emerald-400", bg: "bg-emerald-400/10" },
};

function getInitials(name: string): string {
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((n) => n[0].toUpperCase())
    .join("");
}

export function TaskCard({ task, members, onUpdate, onDelete, isDragging }: TaskCardProps) {
  const { t } = useI18n();
  const [isExpanded, setIsExpanded] = useState(false);
  const [title, setTitle] = useState(task.title);
  const [description, setDescription] = useState(task.description);
  const cardRef = useRef<HTMLDivElement>(null);

  const { attributes, listeners, setNodeRef, transform, transition } = useSortable({
    id: String(task.id),
  });

  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const isDone = task.status === "done";
  const cfg = statusConfig[task.status] || statusConfig.todo;
  const StatusIcon = cfg.icon;

  // Close on click outside
  useEffect(() => {
    if (!isExpanded) return;
    function handleClickOutside(e: MouseEvent) {
      if (cardRef.current && !cardRef.current.contains(e.target as Node)) {
        setIsExpanded(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isExpanded]);

  function handleBlurTitle() {
    if (title !== task.title) onUpdate(task.id, { title });
  }

  function handleBlurDescription() {
    if (description !== task.description) onUpdate(task.id, { description });
  }

  function cycleStatus() {
    const order = ["todo", "in_progress", "done"];
    const next = order[(order.indexOf(task.status) + 1) % order.length];
    onUpdate(task.id, { status: next });
  }

  function handleAssignee(userId: string) {
    onUpdate(task.id, { assigneeId: userId === "" ? null : BigInt(userId) });
  }

  if (isExpanded) {
    return (
      <div
        ref={(node) => {
          setNodeRef(node);
          (cardRef as React.MutableRefObject<HTMLDivElement | null>).current = node;
        }}
        style={style}
        className="rounded-xl border border-primary/20 bg-card/80 shadow-lg backdrop-blur-sm"
      >
        {/* Header */}
        <div className="flex items-center justify-between border-b border-border/30 px-4 py-2.5">
          <div className="flex items-center gap-2">
            <button
              {...attributes}
              {...listeners}
              className="cursor-grab text-muted-foreground/50 hover:text-muted-foreground active:cursor-grabbing"
              tabIndex={-1}
            >
              <GripVertical className="size-3.5" />
            </button>
            <button
              onClick={cycleStatus}
              className={`cursor-pointer rounded-md p-1 ${cfg.bg} transition-colors hover:opacity-80`}
            >
              <StatusIcon className={`size-3.5 ${cfg.color}`} />
            </button>
          </div>
          <button
            onClick={() => setIsExpanded(false)}
            className="cursor-pointer rounded-md p-1 text-muted-foreground/50 hover:bg-accent hover:text-foreground transition-colors"
          >
            <X className="size-3.5" />
          </button>
        </div>

        {/* Body */}
        <div className="px-4 py-3 space-y-3">
          {/* Title */}
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onBlur={handleBlurTitle}
            placeholder={t("tasks.titlePlaceholder")}
            className="w-full bg-transparent text-sm font-medium outline-none placeholder:text-muted-foreground/40"
            autoFocus
          />

          {/* Description */}
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            onBlur={handleBlurDescription}
            placeholder={t("tasks.descriptionPlaceholder")}
            rows={2}
            className="w-full resize-none bg-transparent text-sm text-foreground/70 outline-none placeholder:text-muted-foreground/40"
          />

          {/* Properties row */}
          <div className="flex flex-wrap items-center gap-2 pt-1">
            {/* Assignee */}
            <select
              value={task.assigneeId !== undefined ? String(task.assigneeId) : ""}
              onChange={(e) => handleAssignee(e.target.value)}
              className="cursor-pointer rounded-md border border-border/40 bg-transparent px-2 py-1 text-xs text-foreground/70 outline-none hover:border-border/60 transition-colors"
            >
              <option value="">{t("tasks.unassigned")}</option>
              {members.map((m) => (
                <option key={String(m.userId)} value={String(m.userId)}>
                  {m.userName}
                </option>
              ))}
            </select>

            {/* Spec ref */}
            <input
              value={task.specRef}
              onChange={(e) => onUpdate(task.id, { specRef: e.target.value })}
              placeholder={t("tasks.specRefPlaceholder")}
              className="w-20 rounded-md border border-border/40 bg-transparent px-2 py-1 text-xs text-foreground/70 outline-none placeholder:text-muted-foreground/30 hover:border-border/60 focus:border-primary/50 transition-colors"
            />

            {/* Delete */}
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

  // Collapsed state — Linear style compact card
  return (
    <div
      ref={setNodeRef}
      style={style}
      onClick={() => setIsExpanded(true)}
      {...attributes}
      {...listeners}
      className={`group cursor-pointer rounded-lg border border-border/30 bg-card/50 px-3 py-2.5 transition-all duration-150 hover:border-border/50 hover:bg-card/80 ${
        isDone ? "opacity-60" : ""
      }`}
    >
      <div className="flex items-center gap-2.5">
        {/* Drag indicator */}
        <div className="shrink-0 text-muted-foreground/30 opacity-0 group-hover:opacity-100 transition-opacity">
          <GripVertical className="size-3.5" />
        </div>

        {/* Status icon */}
        <button
          onClick={(e) => {
            e.stopPropagation();
            cycleStatus();
          }}
          className={`shrink-0 cursor-pointer rounded-md p-0.5 ${cfg.bg} transition-colors hover:opacity-80`}
        >
          <StatusIcon className={`size-3.5 ${cfg.color}`} />
        </button>

        {/* Title */}
        <span
          className={`min-w-0 flex-1 truncate text-sm ${
            isDone ? "line-through text-muted-foreground" : "text-foreground/90"
          }`}
        >
          {task.title}
        </span>

        {/* Spec ref badge */}
        {task.specRef && (
          <span className="shrink-0 rounded bg-muted/60 px-1.5 py-0.5 text-[10px] text-muted-foreground">
            {task.specRef}
          </span>
        )}

        {/* Assignee avatar */}
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
    </div>
  );
}
