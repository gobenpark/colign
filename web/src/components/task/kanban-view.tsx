"use client";

import { useState } from "react";
import {
  DndContext,
  DragOverlay,
  closestCorners,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragOverEvent,
  type DragEndEvent,
} from "@dnd-kit/core";
import { arrayMove, sortableKeyboardCoordinates } from "@dnd-kit/sortable";
import { useI18n } from "@/lib/i18n";
import { TaskCard } from "./task-card";
import { KanbanColumn } from "./kanban-column";

type TaskType = {
  id: bigint;
  title: string;
  description: string;
  status: string;
  specRef: string;
  assigneeId?: bigint;
  assigneeName: string;
  orderIndex: number;
};

interface KanbanViewProps {
  tasks: TaskType[];
  members: Array<{ userId: bigint; userName: string }>;
  onCreateTask: (title: string, status: string) => void;
  onUpdateTask: (id: bigint, fields: Record<string, unknown>) => void;
  onDeleteTask: (id: bigint) => void;
  onReorder: (items: Array<{ id: bigint; status: string; orderIndex: number }>) => void;
}

const STATUSES = ["todo", "in_progress", "done"] as const;

function groupByStatus(tasks: TaskType[]): Record<string, TaskType[]> {
  const groups: Record<string, TaskType[]> = { todo: [], in_progress: [], done: [] };
  for (const task of tasks) {
    const bucket = groups[task.status] ?? (groups[task.status] = []);
    bucket.push(task);
  }
  // Sort each bucket by orderIndex
  for (const status of Object.keys(groups)) {
    groups[status].sort((a, b) => a.orderIndex - b.orderIndex);
  }
  return groups;
}

function findContainerOfId(
  grouped: Record<string, TaskType[]>,
  id: string,
): string | undefined {
  // id might be a status column id
  if (id in grouped) return id;
  for (const [status, tasks] of Object.entries(grouped)) {
    if (tasks.some((t) => String(t.id) === id)) return status;
  }
  return undefined;
}

const prefersReducedMotion =
  typeof window !== "undefined" &&
  window.matchMedia("(prefers-reduced-motion: reduce)").matches;

export function KanbanView({
  tasks,
  members,
  onCreateTask,
  onUpdateTask,
  onDeleteTask,
  onReorder,
}: KanbanViewProps) {
  const { t } = useI18n();

  // Local copy of tasks for optimistic drag updates
  const [localTasks, setLocalTasks] = useState<TaskType[]>(tasks);
  const [activeId, setActiveId] = useState<string | null>(null);

  // Sync when parent tasks prop changes (e.g. after server response)
  // We use a simple key comparison to avoid infinite loops
  const tasksKey = tasks.map((t) => `${t.id}:${t.status}:${t.orderIndex}`).join(",");
  const [prevTasksKey, setPrevTasksKey] = useState(tasksKey);
  if (tasksKey !== prevTasksKey) {
    setPrevTasksKey(tasksKey);
    setLocalTasks(tasks);
  }

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 5 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    }),
  );

  const grouped = groupByStatus(localTasks);
  const activeTask = activeId ? localTasks.find((t) => String(t.id) === activeId) : null;

  function handleDragStart(event: DragStartEvent) {
    setActiveId(String(event.active.id));
  }

  function handleDragOver(event: DragOverEvent) {
    const { active, over } = event;
    if (!over) return;

    const activeIdStr = String(active.id);
    const overIdStr = String(over.id);

    const activeContainer = findContainerOfId(grouped, activeIdStr);
    const overContainer = findContainerOfId(grouped, overIdStr);

    if (!activeContainer || !overContainer || activeContainer === overContainer) return;

    // Move the task to the new container optimistically
    setLocalTasks((prev) => {
      return prev.map((task) => {
        if (String(task.id) === activeIdStr) {
          return { ...task, status: overContainer };
        }
        return task;
      });
    });
  }

  function handleDragEnd(event: DragEndEvent) {
    const { active, over } = event;
    setActiveId(null);

    if (!over) return;

    const activeIdStr = String(active.id);
    const overIdStr = String(over.id);

    // Recompute grouped after any status changes from dragOver
    const currentGrouped = groupByStatus(localTasks);

    const activeContainer = findContainerOfId(currentGrouped, activeIdStr);
    const overContainer = findContainerOfId(currentGrouped, overIdStr);

    if (!activeContainer || !overContainer) return;

    let finalTasks = [...localTasks];

    if (activeContainer === overContainer) {
      // Same column: reorder within column
      const columnTasks = [...currentGrouped[activeContainer]];
      const oldIndex = columnTasks.findIndex((t) => String(t.id) === activeIdStr);
      const newIndex = columnTasks.findIndex((t) => String(t.id) === overIdStr);

      if (oldIndex !== -1 && newIndex !== -1 && oldIndex !== newIndex) {
        const reordered = arrayMove(columnTasks, oldIndex, newIndex);
        // Rebuild localTasks with updated orderIndex for this column
        finalTasks = localTasks.map((task) => {
          const idx = reordered.findIndex((r) => r.id === task.id);
          if (idx !== -1) {
            return { ...task, orderIndex: idx };
          }
          return task;
        });
        setLocalTasks(finalTasks);
      }
    }
    // Cross-column moves are already handled in onDragOver via status change.
    // We just need to assign contiguous orderIndex to the destination column.

    // Build the reorder payload for all affected columns.
    // We include all tasks in both source and destination columns.
    const finalGrouped = groupByStatus(finalTasks);

    // Find all columns that have been touched (may include cross-column moves)
    const affectedStatuses = new Set([activeContainer, overContainer]);

    const reorderItems: Array<{ id: bigint; status: string; orderIndex: number }> = [];
    for (const status of affectedStatuses) {
      const col = finalGrouped[status] ?? [];
      col.forEach((task, idx) => {
        reorderItems.push({ id: task.id, status, orderIndex: idx });
      });
    }

    onReorder(reorderItems);
  }

  const columns: Array<{ status: string; label: string; color: string }> = [
    { status: "todo", label: t("tasks.statusTodo"), color: "bg-gray-400" },
    { status: "in_progress", label: t("tasks.statusInProgress"), color: "bg-yellow-500" },
    { status: "done", label: t("tasks.statusDone"), color: "bg-green-500" },
  ];

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragOver={handleDragOver}
      onDragEnd={handleDragEnd}
    >
      <div className="flex gap-4 overflow-x-auto pb-4">
        {columns.map((col) => (
          <KanbanColumn
            key={col.status}
            status={col.status}
            label={col.label}
            color={col.color}
            tasks={grouped[col.status] ?? []}
            members={members}
            onCreateTask={onCreateTask}
            onUpdateTask={onUpdateTask}
            onDeleteTask={onDeleteTask}
          />
        ))}
      </div>

      <DragOverlay
        className="z-50"
        dropAnimation={prefersReducedMotion ? null : undefined}
      >
        {activeTask ? (
          <TaskCard
            task={activeTask}
            members={members}
            onUpdate={onUpdateTask}
            onDelete={onDeleteTask}
            isDragging={true}
          />
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
