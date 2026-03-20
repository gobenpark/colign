"use client";

import { useI18n } from "@/lib/i18n";

export default function SearchPage() {
  const { t } = useI18n();
  return (
    <div className="flex min-h-[60vh] items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-semibold tracking-tight">{t("sidebar.search")}</h1>
        <p className="mt-2 text-sm text-muted-foreground">{t("sidebar.comingSoon")}</p>
      </div>
    </div>
  );
}
