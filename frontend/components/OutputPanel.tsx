"use client";

import { X } from "lucide-react";
import type { RunResult } from "@/lib/types";

interface OutputPanelProps {
  output: RunResult | null;
  onClear: () => void;
}

export default function OutputPanel({ output, onClear }: OutputPanelProps) {
  const exitCode = output?.code ?? null;
  const hasOutput = output && (output.stdout || output.stderr || output.output);

  return (
    <div className="w-96 bg-card border-l border-white/10 flex flex-col">
      <div className="h-14 border-b border-white/10 flex items-center justify-between px-4">
        <div className="flex items-center gap-3">
          <h3 className="text-sm font-semibold text-white">Output</h3>
          {exitCode !== null && (
            <span
              className={exitCode === 0
                  ? "px-2 py-0.5 text-xs font-semibold rounded-full bg-accentGreen/20 text-accentGreen"
                  : "px-2 py-0.5 text-xs font-semibold rounded-full bg-red-500/20 text-red-400"}
            >
              Exit Code: {exitCode}
            </span>
          )}
        </div>
        {hasOutput && (
          <button
            onClick={onClear}
            className="text-muted hover:text-white transition-colors"
            title="Clear output"
            aria-label="Clear output"
          >
            <X size={18} />
          </button>
        )}
      </div>

      <div
        className="flex-1 overflow-auto p-4 font-mono text-sm"
        aria-live="polite"
        aria-atomic="true"
      >
        {!hasOutput ? (
          <div className="text-center text-muted py-8">
            <p>Run your code to see output here</p>
          </div>
        ) : (
          <div className="space-y-3">
            {output.stdout && (
              <div>
                <div className="text-xs text-muted mb-1">STDOUT:</div>
                <pre className="text-white whitespace-pre-wrap">{output.stdout}</pre>
              </div>
            )}
            {output.stderr && (
              <div>
                <div className="text-xs text-red-400 mb-1">STDERR:</div>
                <pre className="text-red-400 whitespace-pre-wrap">{output.stderr}</pre>
              </div>
            )}
            {output.output && !output.stdout && !output.stderr && (
              <div>
                <pre className="text-white whitespace-pre-wrap">{output.output}</pre>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
