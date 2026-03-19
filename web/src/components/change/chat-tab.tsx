"use client";

import { useState, useRef, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useI18n } from "@/lib/i18n";

interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
}

interface ChatTabProps {
  changeId: bigint;
}

export function ChatTab({ changeId }: ChatTabProps) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const { t } = useI18n();
  const [input, setInput] = useState("");
  const [isStreaming, setIsStreaming] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  void changeId;

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  async function handleSend(e: React.FormEvent) {
    e.preventDefault();
    if (!input.trim() || isStreaming) return;

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: "user",
      content: input.trim(),
    };

    setMessages((prev) => [...prev, userMessage]);
    setInput("");
    setIsStreaming(true);

    // Placeholder AI response - will be replaced with real Connect streaming
    const aiMessage: ChatMessage = {
      id: (Date.now() + 1).toString(),
      role: "assistant",
      content: "",
    };
    setMessages((prev) => [...prev, aiMessage]);

    // Simulate streaming for now
    const response = `I'll help you think through this. Based on the SDD methodology, let's break this down:\n\n1. **Requirements**: What specific behavior are you describing?\n2. **Scenarios**: What are the WHEN/THEN conditions?\n3. **Edge cases**: What could go wrong?\n\nCould you elaborate on what you'd like to add to the spec?`;

    for (let i = 0; i < response.length; i++) {
      await new Promise((r) => setTimeout(r, 10));
      setMessages((prev) => {
        const updated = [...prev];
        updated[updated.length - 1] = {
          ...updated[updated.length - 1],
          content: response.slice(0, i + 1),
        };
        return updated;
      });
    }

    setIsStreaming(false);
  }

  return (
    <div className="flex h-full flex-col">
      {/* Messages */}
      <div className="flex-1 overflow-y-auto space-y-4 py-4">
        {messages.length === 0 && (
          <div className="flex h-full items-center justify-center">
            <div className="text-center">
              <p className="text-lg font-medium text-muted-foreground">{t("change.startConversation")}</p>
              <p className="mt-1 text-sm text-muted-foreground/70">
                {t("change.chatDescription")}
              </p>
            </div>
          </div>
        )}
        {messages.map((msg) => (
          <div key={msg.id} className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}>
            <div
              className={`max-w-[80%] rounded-xl px-4 py-3 text-sm ${
                msg.role === "user"
                  ? "bg-primary text-primary-foreground"
                  : "bg-muted text-foreground"
              }`}
            >
              <div className="whitespace-pre-wrap">{msg.content}</div>
              {msg.role === "assistant" && msg.content && !isStreaming && (
                <div className="mt-2 flex gap-1">
                  <Button variant="ghost" size="sm" className="h-6 cursor-pointer px-2 text-xs text-muted-foreground hover:text-foreground">
                    {t("change.applyToSpec")}
                  </Button>
                </div>
              )}
            </div>
          </div>
        ))}
        {isStreaming && (
          <div className="flex justify-start">
            <div className="flex items-center gap-1 px-4 py-2">
              <div className="h-1.5 w-1.5 animate-pulse rounded-full bg-primary" />
              <div className="h-1.5 w-1.5 animate-pulse rounded-full bg-primary [animation-delay:150ms]" />
              <div className="h-1.5 w-1.5 animate-pulse rounded-full bg-primary [animation-delay:300ms]" />
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="border-t border-border/50 pt-4">
        <form onSubmit={handleSend} className="flex gap-2">
          <Input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder={t("change.chatPlaceholder")}
            disabled={isStreaming}
            className="flex-1"
          />
          <Button type="submit" disabled={isStreaming || !input.trim()} className="cursor-pointer">
            {t("common.send")}
          </Button>
        </form>
      </div>
    </div>
  );
}
