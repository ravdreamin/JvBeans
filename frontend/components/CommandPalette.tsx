"use client";

import { useState, useEffect, useRef } from "react";
import Modal from "./Modal";
import { Search } from "lucide-react";

export type Command = {
  id: string;
  label: string;
  hint?: string;
  run: () => void;
};

interface CommandPaletteProps {
  open: boolean;
  onClose: () => void;
  commands: Command[];
}

export default function CommandPalette({ open, onClose, commands }: CommandPaletteProps) {
  const [filter, setFilter] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const filtered = commands.filter((cmd) =>
    cmd.label.toLowerCase().includes(filter.toLowerCase())
  );

  useEffect(() => {
    if (open && inputRef.current) {
      inputRef.current.focus();
    }
  }, [open]);

  useEffect(() => {
    setSelectedIndex(0);
  }, [filter]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((prev) => (prev + 1) % filtered.length);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((prev) => (prev - 1 + filtered.length) % filtered.length);
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (filtered[selectedIndex]) {
        filtered[selectedIndex].run();
        handleClose();
      }
    }
  };

  const handleClose = () => {
    setFilter("");
    setSelectedIndex(0);
    onClose();
  };

  return (
    <Modal open={open} onClose={handleClose} className="max-w-xl">
      <div className="flex flex-col gap-2">
        {/* Search Input */}
        <div className="relative">
          <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-muted" />
          <input
            ref={inputRef}
            type="text"
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Type a command..."
            className="w-full pl-10 pr-4 py-2 bg-black border border-white/10 rounded-lg text-white placeholder:text-muted focus:outline-none focus:border-primaryBlue"
          />
        </div>

        {/* Commands List */}
        <div className="max-h-80 overflow-y-auto">
          {filtered.length === 0 ? (
            <div className="text-center py-8 text-muted text-sm">No commands found</div>
          ) : (
            <div className="flex flex-col gap-1">
              {filtered.map((cmd, index) => (
                <button
                  key={cmd.id}
                  onClick={() => {
                    cmd.run();
                    handleClose();
                  }}
                  onMouseEnter={() => setSelectedIndex(index)}
                  className={`w-full px-4 py-3 text-left rounded-lg transition-colors ${
                    index === selectedIndex
                      ? "bg-primaryBlue/20 border border-primaryBlue/50"
                      : "hover:bg-white/5 border border-transparent"
                  }`}
                >
                  <div className="text-sm font-medium text-white">{cmd.label}</div>
                  {cmd.hint && <div className="text-xs text-muted mt-1">{cmd.hint}</div>}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </Modal>
  );
}
