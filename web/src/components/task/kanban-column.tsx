"use client";

import { useDroppable } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { useI18n } from "@/lib/i18n";
import { TaskCard } from "./task-card";
import { InlineTaskInput } from "./inline-task-input";

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

interface KanbanColumnProps {
  status: string;
  label: string;
  color: string;
  tasks: TaskType[];
  members: Array<{ userId: bigint; userName: string }>;
  onCreateTask: (title: string, status: string) => void;
  onUpdateTask: (id: bigint, fields: Record<string, unknown>) => void;
  onDeleteTask: (id: bigint) => void;
}

export function KanbanColumn({
  status,
  label,
  color,
  tasks,
  members,
  onCreateTask,
  onUpdateTask,
  onDeleteTask,
}: KanbanColumnProps) {
  const { t } = useI18n();
  const { setNodeRef, isOver } = useDroppable({ id: status });

  const itemIds = tasks.map((t) => String(t.id));

  return (
    <div className="flex flex-1 min-w-[250px] flex-col gap-2">
      {/* Column header */}
      <div className="flex items-center gap-2 px-1 py-1.5">
        <span className={`h-2.5 w-2.5 rounded-full flex-shrink-0 ${color}`} />
        <span className="text-sm font-semibold text-foreground">{label}</span>
        <span className="ml-1 rounded-full bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">
          {tasks.length}
        </span>
      </div>

      {/* Droppable task list */}
      <div
        ref={setNodeRef}
        className={[
          "flex flex-col gap-2 rounded-lg p-2 min-h-[200px] transition-colors duration-150",
          isOver ? "bg-muted/60 ring-1 ring-border" : "bg-muted/20",
        ].join(" ")}
      >
        <SortableContext items={itemIds} strategy={verticalListSortingStrategy}>
          {tasks.map((task) => (
            <TaskCard
              key={String(task.id)}
              task={task}
              members={members}
              onUpdate={onUpdateTask}
              onDelete={onDeleteTask}
            />
          ))}
        </SortableContext>
      </div>

      {/* Add task input */}
      <InlineTaskInput onSubmit={(title) => onCreateTask(title, status)} />
    </div>
  );
}
