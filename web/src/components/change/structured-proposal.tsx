"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useI18n } from "@/lib/i18n";
import { documentClient } from "@/lib/document";
import { AcceptanceCriteria } from "@/components/change/acceptance-criteria";
import { ChevronDown, ChevronRight, Sparkles } from "lucide-react";

interface ProposalSections {
  problem: string;
  scope: string;
  outOfScope: string;
  approach: string;
}

const EMPTY_SECTIONS: ProposalSections = {
  problem: "",
  scope: "",
  outOfScope: "",
  approach: "",
};

interface SectionConfig {
  key: keyof ProposalSections;
  i18nKey: string;
  placeholderKey: string;
  required: boolean;
  minRows: number;
}

const SECTIONS: SectionConfig[] = [
  {
    key: "problem",
    i18nKey: "proposal.problem",
    placeholderKey: "proposal.problemPlaceholder",
    required: true,
    minRows: 3,
  },
  {
    key: "scope",
    i18nKey: "proposal.scope",
    placeholderKey: "proposal.scopePlaceholder",
    required: true,
    minRows: 4,
  },
  {
    key: "outOfScope",
    i18nKey: "proposal.outOfScope",
    placeholderKey: "proposal.outOfScopePlaceholder",
    required: false,
    minRows: 2,
  },
  {
    key: "approach",
    i18nKey: "proposal.approach",
    placeholderKey: "proposal.approachPlaceholder",
    required: false,
    minRows: 3,
  },
];

function parseContent(content: string): ProposalSections {
  try {
    const parsed = JSON.parse(content);
    if (parsed && typeof parsed.problem === "string") {
      return { ...EMPTY_SECTIONS, ...parsed };
    }
  } catch {
    // Legacy HTML content — put it all in problem
    if (content.trim()) {
      // Strip HTML tags for plain text
      const text = content
        .replace(/<[^>]*>/g, "")
        .replace(/\s+/g, " ")
        .trim();
      if (
        text &&
        text !== "Explain the motivation for this change. What problem does this solve?"
      ) {
        return { ...EMPTY_SECTIONS, problem: text };
      }
    }
  }
  return { ...EMPTY_SECTIONS };
}

interface StructuredProposalProps {
  changeId: bigint;
  currentStage?: string;
}

export function StructuredProposal({ changeId, currentStage }: StructuredProposalProps) {
  const { t } = useI18n();
  const [sections, setSections] = useState<ProposalSections>(EMPTY_SECTIONS);
  const [loading, setLoading] = useState(true);
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({
    outOfScope: true,
    approach: true,
  });
  const saveTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const sectionsRef = useRef(sections);
  sectionsRef.current = sections;

  // Load
  useEffect(() => {
    async function load() {
      try {
        const res = await documentClient.getDocument({ changeId, type: "proposal" });
        if (res.document?.content) {
          const parsed = parseContent(res.document.content);
          setSections(parsed);
          // Auto-expand optional sections if they have content
          setCollapsed({
            outOfScope: !parsed.outOfScope,
            approach: !parsed.approach,
          });
        }
      } catch {
        // new document
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [changeId]);

  // Auto-save with debounce
  const save = useCallback(() => {
    if (saveTimerRef.current) clearTimeout(saveTimerRef.current);
    saveTimerRef.current = setTimeout(async () => {
      try {
        await documentClient.saveDocument({
          changeId,
          type: "proposal",
          title: "Proposal",
          content: JSON.stringify(sectionsRef.current),
        });
      } catch {
        // save error
      }
    }, 1000);
  }, [changeId]);

  function updateSection(key: keyof ProposalSections, value: string) {
    setSections((prev) => ({ ...prev, [key]: value }));
    save();
  }

  function toggleCollapse(key: string) {
    setCollapsed((prev) => ({ ...prev, [key]: !prev[key] }));
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <div className="h-5 w-5 animate-spin rounded-full border-2 border-primary border-t-transparent" />
      </div>
    );
  }

  const isReviewMode = currentStage === "review" || currentStage === "ready";

  return (
    <div className="py-4 space-y-4">
      {SECTIONS.map((section) => {
        const isCollapsed = collapsed[section.key] && !sections[section.key];
        const hasContent = !!sections[section.key].trim();

        return (
          <div
            key={section.key}
            className="rounded-xl border border-border/40 bg-card/50 transition-colors"
          >
            {/* Section header */}
            <button
              onClick={() => !section.required && toggleCollapse(section.key)}
              className={`flex w-full items-center justify-between px-5 py-3 ${
                section.required ? "" : "cursor-pointer"
              }`}
            >
              <div className="flex items-center gap-2.5">
                {!section.required &&
                  (isCollapsed ? (
                    <ChevronRight className="size-3.5 text-muted-foreground/50" />
                  ) : (
                    <ChevronDown className="size-3.5 text-muted-foreground/50" />
                  ))}
                <span className="text-sm font-medium">{t(section.i18nKey)}</span>
                {section.required ? (
                  <span className="rounded bg-primary/10 px-1.5 py-0.5 text-[10px] font-medium text-primary">
                    {t("proposal.required")}
                  </span>
                ) : (
                  <span className="rounded bg-muted/80 px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
                    {t("proposal.optional")}
                  </span>
                )}
                {hasContent && <div className="h-1.5 w-1.5 rounded-full bg-emerald-400" />}
              </div>
            </button>

            {/* Section body */}
            {(!isCollapsed || section.required) && (
              <div className="border-t border-border/30 px-5 py-4">
                <textarea
                  value={sections[section.key]}
                  onChange={(e) => updateSection(section.key, e.target.value)}
                  placeholder={t(section.placeholderKey)}
                  readOnly={isReviewMode}
                  rows={Math.max(section.minRows, sections[section.key].split("\n").length + 1)}
                  className="w-full resize-none bg-transparent text-sm leading-relaxed text-foreground outline-none placeholder:text-muted-foreground/40"
                />
              </div>
            )}
          </div>
        );
      })}

      {/* AI Generate placeholder — future feature */}
      <div className="flex items-center gap-2 rounded-xl border border-dashed border-border/30 px-5 py-3">
        <Sparkles className="size-4 text-muted-foreground/30" />
        <span className="text-xs text-muted-foreground/40">
          AI-assisted spec generation — coming soon
        </span>
      </div>

      {/* Acceptance Criteria */}
      <AcceptanceCriteria changeId={changeId} reviewMode={isReviewMode} />
    </div>
  );
}
