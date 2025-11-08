"use client";

import { useState, useEffect } from "react";
import { X } from "lucide-react";

interface GenerateModalProps {
  isOpen: boolean;
  onClose: () => void;
  onGenerate: (prompt: string) => void;
}

export default function GenerateModal({ isOpen, onClose, onGenerate }: GenerateModalProps) {
  const [prompt, setPrompt] = useState("");
  const [isGenerating, setIsGenerating] = useState(false);
  const [progress, setProgress] = useState(0);

  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === "Escape" && !isGenerating) {
        onClose();
      }
    };

    if (isOpen) {
      window.addEventListener("keydown", handleEscape);
      return () => window.removeEventListener("keydown", handleEscape);
    }
  }, [isOpen, isGenerating, onClose]);

  if (!isOpen) return null;

  const handleGenerate = async () => {
    if (prompt.trim()) {
      setIsGenerating(true);
      setProgress(0);

      const interval = setInterval(() => {
        setProgress((prev) => (prev >= 90 ? 90 : prev + 10));
      }, 200);

      try {
        await onGenerate(prompt);
        setProgress(100);
        setTimeout(() => {
          setPrompt("");
          setIsGenerating(false);
          setProgress(0);
          clearInterval(interval);
        }, 500);
      } catch (error) {
        setIsGenerating(false);
        setProgress(0);
        clearInterval(interval);
      }
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-70 flex items-center justify-center z-50 p-4">
      <div className="frosted-glass rounded-card shadow-lg max-w-2xl w-full border border-gray-700">
        <div className="px-6 py-4 border-b border-gray-700 flex items-center justify-between">
          <h2 className="text-xl font-bold text-text">Generate Code with AI</h2>
          <button onClick={onClose} className="text-text-secondary hover:text-text" aria-label="Close">
            <X size={20} />
          </button>
        </div>

        <div className="p-6">
          <textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="Type or paste text to generate..."
            className="w-full h-48 px-4 py-3 border border-gray-600 rounded-input bg-surface text-text focus-ring resize-none"
            disabled={isGenerating}
            autoFocus
          />

          {isGenerating && (
            <div className="mt-4">
              <div className="flex items-center justify-between mb-2">
                <span className="text-sm text-text-secondary">Generating...</span>
                <span className="text-sm text-text-secondary">{progress}%</span>
              </div>
              <div className="w-full bg-gray-700 rounded-pill h-2 overflow-hidden">
                <div className="bg-primary h-full transition-all duration-300" style={{ width: `${progress}%` }} />
              </div>
            </div>
          )}

          <p className="mt-4 text-xs text-text-secondary">Rate limit: 20 requests per hour</p>
        </div>

        <div className="px-6 py-4 border-t border-gray-700 flex justify-end gap-3">
          <button
            onClick={onClose}
            disabled={isGenerating}
            className="px-4 py-2 text-sm font-semibold text-text hover:bg-card-hover rounded-input disabled:opacity-50"
          >
            Close
          </button>
          <button
            onClick={handleGenerate}
            disabled={!prompt.trim() || isGenerating}
            className="px-4 py-2 text-sm font-semibold text-white bg-primary rounded-input hover:shadow-glow hover-lift focus-ring disabled:opacity-50"
          >
            {isGenerating ? "Generating..." : "Generate"}
          </button>
        </div>
      </div>
    </div>
  );
}
