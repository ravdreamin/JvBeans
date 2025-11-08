"use client";

import { useState, useEffect, useCallback } from "react";
import ExplorerSidebar from "@/components/ExplorerSidebar";
import EditorHeader from "@/components/EditorHeader";
import OutputPanel from "@/components/OutputPanel";
import Editor from "@monaco-editor/react";
import { getSpaces, getTree, getLog, updateLog, deleteLog, runCode, generateCode } from "@/lib/api";
import { getLanguageFromExtension } from "@/lib/languages";
import type { Space, TreeNode, Log, RunResult } from "@/lib/types";

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
      console.error("Failed to load spaces:", error);
      alert("Failed to load spaces. Is the backend running on http://localhost:8080?");
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
      console.error("Failed to load log:", error);
      alert("Failed to load log");
    }
  };

  const handleSave = useCallback(async () => {
    if (!selectedLog) return;

    setIsSaving(true);
    try {
      await updateLog(selectedLog.id, { code });
      setTimeout(() => setIsSaving(false), 500);
    } catch (error) {
      console.error("Failed to save log:", error);
      alert("Failed to save");
      setIsSaving(false);
    }
  }, [selectedLog, code]);

  const handleRun = async () => {
    if (!selectedLog) return;

    setIsRunning(true);
    try {
      const language = getLanguageFromExtension(selectedLog.name);
      const result = await runCode(language, code);
      setOutput(result);
    } catch (error: any) {
      setOutput({
        stdout: "",
        stderr: error.response?.data?.error || error.message || "Failed to execute code",
        code: 1,
        output: "",
      });
    } finally {
      setIsRunning(false);
    }
  };

  const handleGenerate = async () => {
    const prompt = window.prompt("Describe what code you want to generate:");
    if (!prompt) return;

    setIsGenerating(true);
    try {
      const language = selectedLog ? getLanguageFromExtension(selectedLog.name) : undefined;
      const result = await generateCode(prompt, language);
      setCode(result.code);
    } catch (error: any) {
      alert("Failed to generate code: " + (error.response?.data?.error || error.message));
    } finally {
      setIsGenerating(false);
    }
  };

  const handleDelete = async () => {
    if (!selectedLog) return;
    if (!confirm(`Delete ${selectedLog.name}?`)) return;

    try {
      await deleteLog(selectedLog.id);
      setSelectedLog(null);
      setCode("");
      setOutput(null);
      if (selectedSpace) {
        loadTree(selectedSpace.id);
      }
    } catch (error) {
      console.error("Failed to delete log:", error);
      alert("Failed to delete");
    }
  };

  const handleRefreshTree = () => {
    if (selectedSpace) {
      loadTree(selectedSpace.id);
    }
  };

  // Keyboard shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key === "s") {
        e.preventDefault();
        handleSave();
      } else if ((e.ctrlKey || e.metaKey) && e.key === "r") {
        e.preventDefault();
        handleRun();
      } else if (e.key === "g" && !e.ctrlKey && !e.metaKey && !e.altKey) {
        const target = e.target as HTMLElement;
        if (target.tagName !== "INPUT" && target.tagName !== "TEXTAREA") {
          e.preventDefault();
          handleGenerate();
        }
      } else if (e.key === "Delete" && selectedLog) {
        const target = e.target as HTMLElement;
        if (target.tagName !== "INPUT" && target.tagName !== "TEXTAREA") {
          handleDelete();
        }
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleSave, selectedLog, code]);

  const breadcrumb = selectedLog
    ? `${selectedSpace?.name || ""} / ${selectedLog.path}`
    : selectedSpace?.name || "Select a log";

  const language = selectedLog ? getLanguageFromExtension(selectedLog.name) : "javascript";

  return (
    <div className="h-screen flex flex-col bg-black text-white">
      {/* Header */}
      <EditorHeader
        breadcrumb={breadcrumb}
        language={language}
        onGenerate={handleGenerate}
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
        />

        {/* Editor */}
        <div className="flex-1 border-r border-white/10">
          {selectedLog ? (
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
          ) : (
            <div className="h-full flex items-center justify-center text-muted">
              <div className="text-center">
                <p className="text-xl mb-2">No log selected</p>
                <p className="text-sm">Create or select a log from the explorer</p>
              </div>
            </div>
          )}
        </div>

        {/* Output Panel */}
        <OutputPanel output={output} onClear={() => setOutput(null)} />
      </div>
    </div>
  );
}
