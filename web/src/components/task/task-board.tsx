"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { LayoutGrid, List } from "lucide-react";
import { useI18n } from "@/lib/i18n";
import { taskClient } from "@/lib/task";
import { KanbanView } from "./kanban-view";
import { ListView } from "./list-view";

const STORAGE_KEY = "colign:task-view-mode";

type TaskType = {
  id: bigint;
  changeId: bigint;
  title: string;
  description: string;
  status: string;
  orderIndex: number;
  specRef: string;
  assigneeId?: bigint;
  creatorId?: bigint;
  assigneeName: string;
  creatorName: string;
};

interface TaskBoardProps {
  changeId: bigint;
  members: Array<{ userId: bigint; userName: string }>;
}

type ViewMode = "kanban" | "list";

function readStoredViewMode(): ViewMode {
  if (typeof window === "undefined") return "kanban";
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored === "kanban" || stored === "list") return stored;
  return "kanban";
}

export function TaskBoard({ changeId, members }: TaskBoardProps) {
  const { t } = useI18n();
  const [tasks, setTasks] = useState<TaskType[]>([]);
  const [viewMode, setViewMode] = useState<ViewMode>("kanban");
  const [loading, setLoading] = useState(true);

  type PendingDelete = { task: TaskType };
  const [pendingDelete, setPendingDelete] = useState<PendingDelete | null>(null);
  const undoTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Load persisted view mode on mount
  useEffect(() => {
    setViewMode(readStoredViewMode());
  }, []);

  // Fetch tasks on mount
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    taskClient
      .listTasks({ changeId })
      .then((res) => {
        if (cancelled) return;
        setTasks(
          res.tasks.map((t) => ({
            id: t.id,
            changeId: t.changeId,
            title: t.title,
            description: t.description,
            status: t.status,
            orderIndex: t.orderIndex,
            specRef: t.specRef,
            assigneeId: t.assigneeId,
            creatorId: t.creatorId,
            assigneeName: t.assigneeName,
            creatorName: t.creatorName,
          })),
        );
      })
      .catch((err) => {
        if (!cancelled) console.error("Failed to load tasks:", err);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [changeId]);

  function handleViewModeChange(mode: ViewMode) {
    setViewMode(mode);
    localStorage.setItem(STORAGE_KEY, mode);
  }

  const handleCreate = useCallback(
    async (title: string, status: string) => {
      const tempId = BigInt(-Date.now());
      const tempTask: TaskType = {
        id: tempId,
        changeId,
        title,
        description: "",
        status,
        orderIndex: 0,
        specRef: "",
        assigneeName: "",
        creatorName: "",
      };

      setTasks((prev) => [...prev, tempTask]);

      try {
        const res = await taskClient.createTask({ changeId, title, status });
        if (!res.task) throw new Error("No task in response");
        const serverTask = res.task;
        setTasks((prev) =>
          prev.map((t) =>
            t.id === tempId
              ? {
                  id: serverTask.id,
                  changeId: serverTask.changeId,
                  title: serverTask.title,
                  description: serverTask.description,
                  status: serverTask.status,
                  orderIndex: serverTask.orderIndex,
                  specRef: serverTask.specRef,
                  assigneeId: serverTask.assigneeId,
                  creatorId: serverTask.creatorId,
                  assigneeName: serverTask.assigneeName,
                  creatorName: serverTask.creatorName,
                }
              : t,
          ),
        );
      } catch (err) {
        console.error("Failed to create task:", err);
        setTasks((prev) => prev.filter((t) => t.id !== tempId));
      }
    },
    [changeId],
  );

  const handleUpdate = useCallback(async (id: bigint, fields: Record<string, unknown>) => {
    // Snapshot for rollback
    let prevTasks: TaskType[] = [];
    setTasks((prev) => {
      prevTasks = prev;
      return prev.map((t) => (t.id === id ? { ...t, ...(fields as Partial<TaskType>) } : t));
    });

    const req: {
      id: bigint;
      title?: string;
      description?: string;
      status?: string;
      specRef?: string;
      assigneeId?: bigint;
      clearAssignee?: boolean;
    } = { id };

    if ("title" in fields) req.title = fields.title as string;
    if ("description" in fields) req.description = fields.description as string;
    if ("status" in fields) req.status = fields.status as string;
    if ("specRef" in fields) req.specRef = fields.specRef as string;
    if ("assigneeId" in fields) {
      const val = fields.assigneeId;
      if (val == null) {
        req.clearAssignee = true;
      } else {
        req.assigneeId = val as bigint;
      }
    }

    try {
      await taskClient.updateTask(req);
    } catch (err) {
      console.error("Failed to update task:", err);
      setTasks(prevTasks);
    }
  }, []);

  const handleUndo = useCallback(() => {
    if (!pendingDelete) return;
    if (undoTimeoutRef.current !== null) {
      clearTimeout(undoTimeoutRef.current);
      undoTimeoutRef.current = null;
    }
    setTasks((prev) => [...prev, pendingDelete.task]);
    setPendingDelete(null);
  }, [pendingDelete]);

  const handleDelete = useCallback(
    (id: bigint) => {
      // Cancel any previous pending delete
      if (undoTimeoutRef.current !== null) {
        clearTimeout(undoTimeoutRef.current);
        undoTimeoutRef.current = null;
      }
      // If there was a previous pending delete, commit it immediately
      if (pendingDelete) {
        taskClient.deleteTask({ id: pendingDelete.task.id }).catch((err) => {
          console.error("Failed to delete task:", err);
        });
        setPendingDelete(null);
      }

      const taskToDelete = tasks.find((t) => t.id === id);
      if (!taskToDelete) return;

      setTasks((prev) => prev.filter((t) => t.id !== id));
      setPendingDelete({ task: taskToDelete });

      undoTimeoutRef.current = setTimeout(() => {
        undoTimeoutRef.current = null;
        setPendingDelete(null);
        taskClient.deleteTask({ id }).catch((err) => {
          console.error("Failed to delete task:", err);
        });
      }, 3000);
    },
    [tasks, pendingDelete],
  );

  const handleReorder = useCallback(
    async (items: Array<{ id: bigint; status: string; orderIndex: number }>) => {
      try {
        await taskClient.reorderTasks({
          changeId,
          items: items.map((i) => ({
            id: i.id,
            status: i.status,
            orderIndex: i.orderIndex,
          })),
        });
      } catch (err) {
        console.error("Failed to reorder tasks:", err);
        // Re-fetch to reset state
        try {
          const res = await taskClient.listTasks({ changeId });
          setTasks(
            res.tasks.map((t) => ({
              id: t.id,
              changeId: t.changeId,
              title: t.title,
              description: t.description,
              status: t.status,
              orderIndex: t.orderIndex,
              specRef: t.specRef,
              assigneeId: t.assigneeId,
              creatorId: t.creatorId,
              assigneeName: t.assigneeName,
              creatorName: t.creatorName,
            })),
          );
        } catch (fetchErr) {
          console.error("Failed to re-fetch tasks after reorder error:", fetchErr);
        }
      }
    },
    [changeId],
  );

  return (
    <div className="flex flex-col gap-4">
      {/* Toolbar */}
      <div className="flex items-center gap-1">
        <button
          onClick={() => handleViewModeChange("kanban")}
          className={`inline-flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm transition-colors ${
            viewMode === "kanban"
              ? "bg-muted text-foreground"
              : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
          }`}
          aria-label={t("tasks.viewKanban")}
        >
          <LayoutGrid className="h-4 w-4" />
          <span>{t("tasks.viewKanban")}</span>
        </button>
        <button
          onClick={() => handleViewModeChange("list")}
          className={`inline-flex items-center gap-1.5 rounded-md px-2.5 py-1.5 text-sm transition-colors ${
            viewMode === "list"
              ? "bg-muted text-foreground"
              : "text-muted-foreground hover:bg-muted/50 hover:text-foreground"
          }`}
          aria-label={t("tasks.viewList")}
        >
          <List className="h-4 w-4" />
          <span>{t("tasks.viewList")}</span>
        </button>
      </div>

      {/* Content */}
      {loading ? (
        <div className="space-y-3">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-16 animate-pulse rounded-md bg-muted" />
          ))}
        </div>
      ) : viewMode === "kanban" ? (
        <KanbanView
          tasks={tasks}
          members={members}
          onCreateTask={handleCreate}
          onUpdateTask={handleUpdate}
          onDeleteTask={handleDelete}
          onReorder={handleReorder}
        />
      ) : (
        <ListView
          tasks={tasks}
          members={members}
          onCreateTask={handleCreate}
          onUpdateTask={handleUpdate}
          onDeleteTask={handleDelete}
        />
      )}

      {/* Delete undo toast */}
      {pendingDelete && (
        <div className="fixed bottom-4 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 rounded-lg bg-foreground px-4 py-2.5 text-background shadow-lg">
          <span className="text-sm">{t("tasks.deleteUndo")}</span>
          <button onClick={handleUndo} className="cursor-pointer text-sm font-medium underline">
            {t("tasks.undo")}
          </button>
        </div>
      )}
    </div>
  );
}
