"use client";

import { useState, useEffect, useCallback } from "react";
import ExplorerSidebar from "@/components/ExplorerSidebar";
import EditorHeader from "@/components/EditorHeader";
import OutputPanel from "@/components/OutputPanel";
import CommandPalette, { type Command } from "@/components/CommandPalette";
import Editor from "@monaco-editor/react";
import { getSpaces, getTree, getLog, updateLog, deleteLog, runCode, generateCode, ApiError } from "@/lib/api";
import { getLanguageFromExtension } from "@/lib/languages";
import type { Space, TreeNode, Log, RunResult } from "@/lib/types";
import { useToast } from "@/components/Toast";

export default function FlowPage() {
  const [spaces, setSpaces] = useState<Space[]>([]);
  const [selectedSpace, setSelectedSpace] = useState<Space | null>(null);
  const [tree, setTree] = useState<TreeNode[]>([]);
  const [selectedLog, setSelectedLog] = useState<Log | null>(null);
  const [code, setCode] = useState("");
  const [output, setOutput] = useState<RunResult | null>(null);
  const [isRunning, setIsRunning] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const [generatePrompt, setGeneratePrompt] = useState("");
  const [showGenerateInput, setShowGenerateInput] = useState(false);

  const { showToast } = useToast();

  // Load spaces on mount
  useEffect(() => {
    loadSpaces();
  }, []);

  // Load tree when space changes
  useEffect(() => {
    if (selectedSpace) {
      loadTree(selectedSpace.id);
    } else {
      setTree([]);
    }
  }, [selectedSpace]);

  const loadSpaces = async () => {
    try {
      const data = await getSpaces();
      setSpaces(data);
      if (data.length > 0 && !selectedSpace) {
        setSelectedSpace(data[0]);
      }
    } catch (error) {
      const err = error as ApiError;
      showToast(err.message || "Failed to load spaces. Is the backend running?", "error");
    }
  };

  const loadTree = async (spaceId: string) => {
    try {
      const data = await getTree(spaceId);
      setTree(data);
    } catch (error) {
      console.error("Failed to load tree:", error);
      setTree([]);
    }
  };

  const handleSelectLog = async (logId: string) => {
    try {
      const log = await getLog(logId);
      setSelectedLog(log);
      setCode(log.code);
      setOutput(null);
    } catch (error) {
      const err = error as ApiError;
      showToast(err.message || "Failed to load log", "error");
    }
  };

  const handleSave = useCallback(async () => {
    if (!selectedLog) return;

    setIsSaving(true);
    try {
      await updateLog(selectedLog.id, { code });
      showToast("Saved successfully", "success");
      setTimeout(() => setIsSaving(false), 500);
    } catch (error) {
      const err = error as ApiError;
      showToast(err.message || "Failed to save", "error");
      setIsSaving(false);
    }
  }, [selectedLog, code, showToast]);

  const handleRun = async () => {
    if (!selectedLog) return;

    setIsRunning(true);
    try {
      const language = getLanguageFromExtension(selectedLog.name);
      const result = await runCode(language, code);
      setOutput(result);
      showToast("Code executed successfully", "success");
    } catch (error) {
      const err = error as ApiError;
      setOutput({
        stdout: "",
        stderr: err.message || "Failed to execute code",
        code: 1,
        output: "",
      });
      showToast(err.message || "Execution failed", "error");
    } finally {
      setIsRunning(false);
    }
  };

  // UPDATED: AI Generate with proper error handling for API keys
  const handleGenerate = async () => {
    if (!selectedLog && !showGenerateInput) {
      setShowGenerateInput(true);
      return;
    }

    const prompt = generatePrompt.trim();
    if (!prompt) {
      showToast("Please enter a prompt", "info");
      return;
    }

    setIsGenerating(true);
    setShowGenerateInput(false);
    try {
      const language = selectedLog ? getLanguageFromExtension(selectedLog.name) : undefined;
      const result = await generateCode(prompt, language);
      setCode(result.code);
      showToast("Code generated successfully", "success");
      setGeneratePrompt("");
    } catch (error) {
      const err = error as ApiError;

      // UPDATED: Check for API key errors and show non-blocking toast
      if (err.code === "API_KEY_INVALID" || err.message?.includes("API key")) {
        showToast(
          "AI provider not configured. Add OpenAI or Gemini API keys in backend .env to enable generation.",
          "error"
        );
      } else {
        showToast(err.message || "Failed to generate code", "error");
      }
    } finally {
      setIsGenerating(false);
    }
  };

  const handleDelete = async () => {
    if (!selectedLog) return;

    try {
      await deleteLog(selectedLog.id);
      showToast("Log deleted successfully", "success");
      setSelectedLog(null);
      setCode("");
      setOutput(null);
      if (selectedSpace) {
        loadTree(selectedSpace.id);
      }
    } catch (error) {
      const err = error as ApiError;
      showToast(err.message || "Failed to delete", "error");
    }
  };

  const handleRefreshTree = () => {
    if (selectedSpace) {
      loadTree(selectedSpace.id);
    }
  };

  // UPDATED: Keyboard shortcuts with command palette support
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Command Palette: Ctrl/Cmd+Shift+P
      if ((e.ctrlKey || e.metaKey) && e.shiftKey && e.key === "P") {
        e.preventDefault();
        setCommandPaletteOpen(true);
        return;
      }

      // Save: Ctrl/Cmd+S
      if ((e.ctrlKey || e.metaKey) && e.key === "s") {
        e.preventDefault();
        handleSave();
        return;
      }

      // Run: Ctrl/Cmd+R
      if ((e.ctrlKey || e.metaKey) && e.key === "r") {
        e.preventDefault();
        handleRun();
        return;
      }

      // Only allow these shortcuts when not typing in input/textarea
      const target = e.target as HTMLElement;
      const isTyping = target.tagName === "INPUT" || target.tagName === "TEXTAREA";

      // Generate: G key
      if (e.key === "g" && !isTyping && !e.ctrlKey && !e.metaKey && !e.altKey) {
        e.preventDefault();
        setShowGenerateInput(true);
        return;
      }

      // Delete: Delete key
      if (e.key === "Delete" && selectedLog && !isTyping) {
        e.preventDefault();
        handleDelete();
        return;
      }

      // Create Vault: A key
      if (e.key === "a" && !isTyping && !e.ctrlKey && !e.metaKey && !e.altKey) {
        e.preventDefault();
        // This will be handled by ExplorerSidebar
        return;
      }

      // Create Log: N key
      if (e.key === "n" && !isTyping && !e.ctrlKey && !e.metaKey && !e.altKey) {
        e.preventDefault();
        // This will be handled by ExplorerSidebar
        return;
      }

      // Rename: F2 key
      if (e.key === "F2" && !isTyping) {
        e.preventDefault();
        // This will be handled by ExplorerSidebar
        return;
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleSave, selectedLog, handleRun]);

  // Command palette commands
  const commands: Command[] = [
    {
      id: "create-space",
      label: "Create Space",
      hint: "Create a new workspace",
      run: () => {
        // This will be handled by ExplorerSidebar's inline input
        showToast("Click the + icon next to 'Space' to create a new space", "info");
      },
    },
    {
      id: "create-vault",
      label: "Create Vault in current Space",
      hint: "Create a new folder (A)",
      run: () => {
        showToast("Right-click in the explorer to create a vault", "info");
      },
    },
    {
      id: "create-log",
      label: "Create Log in selected Vault",
      hint: "Create a new code file (N)",
      run: () => {
        showToast("Right-click on a vault to create a log", "info");
      },
    },
    {
      id: "save",
      label: "Save",
      hint: "Save current file (Ctrl/Cmd+S)",
      run: handleSave,
    },
    {
      id: "run",
      label: "Run Code",
      hint: "Execute current file (Ctrl/Cmd+R)",
      run: handleRun,
    },
    {
      id: "generate",
      label: "Generate Code",
      hint: "AI code generation (G)",
      run: () => setShowGenerateInput(true),
    },
    {
      id: "delete",
      label: "Delete Selected",
      hint: "Delete current file (Del)",
      run: handleDelete,
    },
  ];

  const breadcrumb = selectedLog
    ? `${selectedSpace?.name || ""} / ${selectedLog.path || selectedLog.name}`
    : selectedSpace?.name || "Select a log";

  const language = selectedLog ? getLanguageFromExtension(selectedLog.name) : "javascript";

  return (
    <div className="h-screen flex flex-col bg-black text-white">
      {/* Header */}
      <EditorHeader
        breadcrumb={breadcrumb}
        language={language}
        onGenerate={() => setShowGenerateInput(true)}
        onRun={handleRun}
        onSave={handleSave}
        onDelete={handleDelete}
        isGenerating={isGenerating}
        isRunning={isRunning}
        isSaving={isSaving}
        hasSelectedLog={!!selectedLog}
      />

      {/* Main Layout: Explorer | Editor | Output */}
      <div className="flex-1 flex overflow-hidden">
        {/* Explorer Sidebar */}
        <ExplorerSidebar
          spaces={spaces}
          selectedSpace={selectedSpace}
          tree={tree}
          onSelectSpace={setSelectedSpace}
          onSelectLog={handleSelectLog}
          onRefresh={handleRefreshTree}
          onSpacesChange={loadSpaces}
          onOpenCommandPalette={() => setCommandPaletteOpen(true)}
        />

        {/* Editor */}
        <div className="flex-1 border-r border-white/10 relative">
          {selectedLog ? (
            <>
              <Editor
                height="100%"
                language={language}
                value={code}
                onChange={(value) => setCode(value || "")}
                theme="vs-dark"
                options={{
                  minimap: { enabled: false },
                  fontSize: 14,
                  fontFamily: "'Fira Code', 'Consolas', 'Monaco', monospace",
                  lineNumbers: "on",
                  rulers: [],
                  scrollBeyondLastLine: false,
                  automaticLayout: true,
                }}
              />

              {/* Generate Prompt Input - Floating */}
              {showGenerateInput && (
                <div className="absolute inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-10">
                  <div className="bg-card border border-white/10 rounded-lg p-6 max-w-xl w-full mx-4">
                    <h3 className="text-lg font-semibold text-white mb-4">Generate Code</h3>
                    <input
                      type="text"
                      value={generatePrompt}
                      onChange={(e) => setGeneratePrompt(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") {
                          handleGenerate();
                        } else if (e.key === "Escape") {
                          setShowGenerateInput(false);
                          setGeneratePrompt("");
                        }
                      }}
                      placeholder="Describe what code you want to generate..."
                      autoFocus
                      className="w-full px-4 py-2 bg-black border border-white/10 rounded-lg text-white placeholder:text-muted focus:outline-none focus:border-primaryBlue mb-4"
                    />
                    <div className="flex gap-2 justify-end">
                      <button
                        onClick={() => {
                          setShowGenerateInput(false);
                          setGeneratePrompt("");
                        }}
                        className="px-4 py-2 text-sm text-muted hover:text-white transition-colors"
                      >
                        Cancel (Esc)
                      </button>
                      <button
                        onClick={handleGenerate}
                        disabled={!generatePrompt.trim() || isGenerating}
                        className="px-4 py-2 text-sm font-semibold text-white bg-primaryBlue rounded-lg hover:bg-primaryBlue/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                      >
                        {isGenerating ? "Generating..." : "Generate (Enter)"}
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </>
          ) : (
            <div className="h-full flex items-center justify-center text-muted">
              <div className="text-center">
                <p className="text-xl mb-2">No log selected</p>
                <p className="text-sm">Create or select a log from the explorer</p>
                <p className="text-xs mt-4">Press Ctrl+Shift+P for commands</p>
              </div>
            </div>
          )}
        </div>

        {/* Output Panel */}
        <OutputPanel output={output} onClear={() => setOutput(null)} />
      </div>

      {/* Command Palette */}
      <CommandPalette
        open={commandPaletteOpen}
        onClose={() => setCommandPaletteOpen(false)}
        commands={commands}
      />
    </div>
  );
}
