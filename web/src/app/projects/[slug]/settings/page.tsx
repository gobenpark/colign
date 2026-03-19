"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Header } from "@/components/layout/header";
import { workflowClient } from "@/lib/workflow";

type SettingsTab = "general" | "members" | "approval" | "danger";

const tabs: { id: SettingsTab; label: string }[] = [
  { id: "general", label: "General" },
  { id: "members", label: "Members" },
  { id: "approval", label: "Approval Policy" },
  { id: "danger", label: "Danger Zone" },
];

const policyOptions = [
  { value: "owner_one", label: "Owner 1 approval" },
  { value: "editor_two", label: "2+ Editor approvals" },
  { value: "all", label: "All members approve" },
  { value: "auto_pass", label: "Auto-pass (no approval)" },
];

export default function ProjectSettingsPage() {
  const params = useParams();
  const slug = params.slug as string;
  const router = useRouter();

  const [activeTab, setActiveTab] = useState<SettingsTab>("general");

  // General
  const [projectName, setProjectName] = useState("");
  const [projectDescription, setProjectDescription] = useState("");

  // Members
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("editor");

  // Approval
  const [policy, setPolicy] = useState("owner_one");
  const [minCount, setMinCount] = useState(1);

  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState("");

  function showSaved(section: string) {
    setSaved(section);
    setTimeout(() => setSaved(""), 2000);
  }

  async function handleSave(section: string) {
    setSaving(true);
    if (section === "approval") {
      try {
        await workflowClient.setApprovalPolicy({
          projectId: BigInt(1), // TODO: from context
          policy,
          minCount,
        });
      } catch {
        // handle error
      }
    }
    // TODO: other API calls
    await new Promise((r) => setTimeout(r, 300));
    setSaving(false);
    showSaved(section);
  }

  return (
    <div className="min-h-screen">
      <Header
        breadcrumbs={[{ label: slug, href: `/projects/${slug}` }, { label: "Settings" }]}
      />

      <div className="mx-auto flex max-w-5xl gap-8 px-6 py-8">
        {/* Sidebar */}
        <nav className="w-48 shrink-0">
          <ul className="space-y-1">
            {tabs.map((tab) => (
              <li key={tab.id}>
                <button
                  onClick={() => setActiveTab(tab.id)}
                  className={`w-full cursor-pointer rounded-lg px-3 py-2 text-left text-sm transition-colors duration-200 ${
                    activeTab === tab.id
                      ? "bg-accent text-foreground"
                      : "text-muted-foreground hover:bg-accent/50 hover:text-foreground"
                  } ${tab.id === "danger" ? "text-destructive" : ""}`}
                >
                  {tab.label}
                </button>
              </li>
            ))}
          </ul>
        </nav>

        {/* Content */}
        <div className="flex-1 space-y-6">
          {/* General */}
          {activeTab === "general" && (
            <Card className="border-border/50">
              <CardHeader>
                <CardTitle>General</CardTitle>
                <CardDescription>Project name and description</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="project-name">Project Name</Label>
                  <Input
                    id="project-name"
                    value={projectName}
                    onChange={(e) => setProjectName(e.target.value)}
                    placeholder="My App"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="project-desc">Description</Label>
                  <Input
                    id="project-desc"
                    value={projectDescription}
                    onChange={(e) => setProjectDescription(e.target.value)}
                    placeholder="A brief description"
                  />
                </div>
                <div className="flex items-center gap-3 pt-2">
                  <Button
                    onClick={() => handleSave("general")}
                    disabled={saving}
                    className="cursor-pointer"
                  >
                    {saving ? "Saving..." : "Save"}
                  </Button>
                  {saved === "general" && <span className="text-sm text-emerald-400">Saved</span>}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Members */}
          {activeTab === "members" && (
            <>
              <Card className="border-border/50">
                <CardHeader>
                  <CardTitle>Invite Member</CardTitle>
                  <CardDescription>Add team members by email</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex gap-2">
                    <Input
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                      placeholder="email@example.com"
                      type="email"
                      className="flex-1"
                    />
                    <Select value={inviteRole} onValueChange={setInviteRole}>
                      <SelectTrigger className="w-32 cursor-pointer">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="editor" className="cursor-pointer">
                          Editor
                        </SelectItem>
                        <SelectItem value="viewer" className="cursor-pointer">
                          Viewer
                        </SelectItem>
                      </SelectContent>
                    </Select>
                    <Button
                      onClick={() => handleSave("invite")}
                      disabled={saving || !inviteEmail}
                      className="cursor-pointer"
                    >
                      Invite
                    </Button>
                  </div>
                  {saved === "invite" && (
                    <p className="mt-2 text-sm text-emerald-400">Invitation sent</p>
                  )}
                </CardContent>
              </Card>

              <Card className="border-border/50">
                <CardHeader>
                  <CardTitle>Members</CardTitle>
                  <CardDescription>People with access to this project</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {/* Placeholder member list */}
                    <div className="flex items-center justify-between rounded-lg border border-border/50 p-3">
                      <div className="flex items-center gap-3">
                        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
                          BP
                        </div>
                        <div>
                          <p className="text-sm font-medium">Ben Park</p>
                          <p className="text-xs text-muted-foreground">ben@example.com</p>
                        </div>
                      </div>
                      <span className="rounded-full bg-primary/10 px-2.5 py-0.5 text-xs font-medium text-primary">
                        Owner
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </>
          )}

          {/* Approval Policy */}
          {activeTab === "approval" && (
            <Card className="border-border/50">
              <CardHeader>
                <CardTitle>Approval Policy</CardTitle>
                <CardDescription>
                  Configure how changes are approved before implementation
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-5">
                <div className="space-y-2">
                  <Label>Policy Type</Label>
                  <Select value={policy} onValueChange={setPolicy}>
                    <SelectTrigger className="cursor-pointer">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {policyOptions.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value} className="cursor-pointer">
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                {policy !== "auto_pass" && policy !== "all" && (
                  <div className="space-y-2">
                    <Label>Minimum Approvals</Label>
                    <Input
                      type="number"
                      min={1}
                      max={10}
                      value={minCount}
                      onChange={(e) => setMinCount(Number(e.target.value))}
                    />
                  </div>
                )}
                <div className="flex items-center gap-3 pt-2">
                  <Button
                    onClick={() => handleSave("approval")}
                    disabled={saving}
                    className="cursor-pointer"
                  >
                    {saving ? "Saving..." : "Save Policy"}
                  </Button>
                  {saved === "approval" && <span className="text-sm text-emerald-400">Saved</span>}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Danger Zone */}
          {activeTab === "danger" && (
            <Card className="border-destructive/30">
              <CardHeader>
                <CardTitle className="text-destructive">Danger Zone</CardTitle>
                <CardDescription>These actions are irreversible</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center justify-between rounded-lg border border-destructive/20 p-4">
                  <div>
                    <p className="text-sm font-medium">Delete this project</p>
                    <p className="text-xs text-muted-foreground">
                      All changes, specs, and data will be permanently deleted
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    className="cursor-pointer border-destructive/50 text-destructive hover:bg-destructive/10"
                    onClick={() => {
                      if (confirm("Are you sure? This cannot be undone.")) {
                        // TODO: API call
                        router.push("/projects");
                      }
                    }}
                  >
                    Delete Project
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
