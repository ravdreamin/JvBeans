"use client";

import { Sparkles, Play, Save, Trash2 } from "lucide-react";
import { getLanguageLabel, getLanguageColor } from "@/lib/languages";

interface EditorHeaderProps {
  breadcrumb: string;
  language: string;
  onGenerate: () => void;
  onRun: () => void;
  onSave: () => void;
  onDelete: () => void;
  isGenerating: boolean;
  isRunning: boolean;
  isSaving: boolean;
  hasSelectedLog: boolean;
}

export default function EditorHeader({
  breadcrumb,
  language,
  onGenerate,
  onRun,
  onSave,
  onDelete,
  isGenerating,
  isRunning,
  isSaving,
  hasSelectedLog,
}: EditorHeaderProps) {
  const languageColor = getLanguageColor(language);

  return (
    <div className="h-14 bg-card border-b border-white/10 flex items-center justify-between px-4">
      {/* Left: Breadcrumb only (no language dropdown) */}
      <div className="flex items-center gap-3">
        <div className="text-sm text-white font-medium truncate max-w-md">
          {breadcrumb}
        </div>
        {hasSelectedLog && (
          <span
            className="px-2 py-0.5 text-xs font-semibold rounded-full"
            style={{
              backgroundColor: `${languageColor}20`,
              color: languageColor,
            }}
          >
            {getLanguageLabel(language)}
          </span>
        )}
      </div>

      {/* Right: Action buttons */}
      <div className="flex items-center gap-2">
        <button
          onClick={onGenerate}
          disabled={isGenerating || !hasSelectedLog}
          className="flex items-center gap-2 px-4 py-1.5 text-sm font-semibold text-white bg-accentCyan/20 border border-accentCyan/30 rounded-full hover:bg-accentCyan/30 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          title="Generate with AI (G)"
        >
          <Sparkles size={16} />
          {isGenerating ? "Generating..." : "Generate"}
        </button>

        <button
          onClick={onRun}
          disabled={isRunning || !hasSelectedLog}
          className="flex items-center gap-2 px-4 py-1.5 text-sm font-semibold text-white bg-primaryBlue/20 border border-primaryBlue/30 rounded-full hover:bg-primaryBlue/30 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          title="Run code (Ctrl/Cmd+R)"
        >
          <Play size={16} />
          {isRunning ? "Running..." : "Run"}
        </button>

        <button
          onClick={onSave}
          disabled={isSaving || !hasSelectedLog}
          className="flex items-center gap-2 px-4 py-1.5 text-sm font-semibold text-white bg-primaryTeal/20 border border-primaryTeal/30 rounded-full hover:bg-primaryTeal/30 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          title="Save (Ctrl/Cmd+S)"
        >
          <Save size={16} />
          {isSaving ? "Saved!" : "Save"}
        </button>

        <button
          onClick={onDelete}
          disabled={!hasSelectedLog}
          className="flex items-center gap-2 px-4 py-1.5 text-sm font-semibold text-white bg-red-500/20 border border-red-500/30 rounded-full hover:bg-red-500/30 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
          title="Delete (Del)"
        >
          <Trash2 size={16} />
          Delete
        </button>
      </div>
    </div>
  );
}
