"use client";

import { useState } from "react";
import { ChevronRight, ChevronDown, Folder, File, Plus, MoreVertical } from "lucide-react";
import { createSpace, createVault, createLog, updateVault, deleteVault, updateLog, deleteLog as deleteLogAPI, ApiError } from "@/lib/api";
import type { Space, TreeNode } from "@/lib/types";
import InlineNameInput from "./InlineNameInput";
import { useToast } from "./Toast";

interface ExplorerSidebarProps {
  spaces: Space[];
  selectedSpace: Space | null;
  tree: TreeNode[];
  onSelectSpace: (space: Space) => void;
  onSelectLog: (logId: string) => void;
  onRefresh: () => void;
  onSpacesChange: () => void;
  onOpenCommandPalette: () => void;
}

type InlineMode =
  | { type: "create-space" }
  | { type: "create-vault"; parentId?: string }
  | { type: "create-log"; vaultId: string }
  | { type: "rename"; node: TreeNode }
  | null;

export default function ExplorerSidebar({
  spaces,
  selectedSpace,
  tree,
  onSelectSpace,
  onSelectLog,
  onRefresh,
  onSpacesChange,
  onOpenCommandPalette,
}: ExplorerSidebarProps) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; node: TreeNode } | null>(null);
  const [inlineMode, setInlineMode] = useState<InlineMode>(null);
  const { showToast } = useToast();

  const toggleExpand = (nodeId: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  };

  // Validation
  const validateSpaceOrVaultName = (name: string): string | null => {
    if (!/^[A-Za-z0-9 _.-]{1,64}$/.test(name)) {
      return "Name must be 1-64 characters: letters, numbers, spaces, _, -, .";
    }
    return null;
  };

  const validateLogFilename = (filename: string): string | null => {
    if (!/\.[a-z0-9]+$/i.test(filename)) {
      return "Filename must include an extension (e.g., .js, .py)";
    }
    if (/[/\\:*?"<>|]/.test(filename)) {
      return 'Filename cannot contain: / \\ : * ? " < > |';
    }
    return null;
  };

  // CRUD Handlers
  const handleCreateSpace = async (name: string) => {
    try {
      await createSpace(name);
      showToast("Space created successfully", "success");
      onSpacesChange();
      setInlineMode(null);
    } catch (error) {
      const err = error as ApiError;
      showToast(err.message || "Failed to create space", "error");
      throw error;
    }
  };

  const handleCreateVault = async (name: string, parentId?: string) => {
    if (!selectedSpace) return;
    try {
      await createVault(selectedSpace.id, name, parentId);
      showToast("Vault created successfully", "success");
      onRefresh();
      setInlineMode(null);
    } catch (error) {
      const err = error as ApiError;
      showToast(err.message || "Failed to create vault", "error");
      throw error;
    }
  };

  const handleCreateLog = async (filename: string, vaultId: string) => {
    if (!selectedSpace) return;
    try {
      await createLog(selectedSpace.id, vaultId, filename);
      showToast("Log created successfully", "success");
      onRefresh();
      setInlineMode(null);
    } catch (error) {
      const err = error as ApiError;
      showToast(err.message || "Failed to create log", "error");
      throw error;
    }
  };

  const handleRename = async (name: string, node: TreeNode) => {
    try {
      if (node.type === "vault") {
        await updateVault(node.id, name);
        showToast("Vault renamed successfully", "success");
      } else if (node.type === "log") {
        await updateLog(node.id, { name });
        showToast("Log renamed successfully", "success");
      }
      onRefresh();
      setInlineMode(null);
    } catch (error) {
      const err = error as ApiError;
      showToast(err.message || "Failed to rename", "error");
      throw error;
    }
  };

  const handleDelete = async (node: TreeNode) => {
    try {
      if (node.type === "vault") {
        await deleteVault(node.id);
        showToast("Vault deleted successfully", "success");
      } else if (node.type === "log") {
        await deleteLogAPI(node.id);
        showToast("Log deleted successfully", "success");
      }
      onRefresh();
    } catch (error) {
      const err = error as ApiError;
      showToast(err.message || "Failed to delete", "error");
    }
  };

  // Context menu
  const handleContextMenu = (e: React.MouseEvent, node: TreeNode) => {
    e.preventDefault();
    e.stopPropagation();
    setContextMenu({ x: e.clientX, y: e.clientY, node });
  };

  const closeContextMenu = () => {
    setContextMenu(null);
  };

  // Render tree nodes
  const renderNode = (node: TreeNode, depth: number = 0) => {
    const isExpanded = expanded.has(node.id);
    const isVault = node.type === "vault";
    const isLog = node.type === "log";

    // Show inline rename input for this node
    if (inlineMode?.type === "rename" && inlineMode.node.id === node.id) {
      return (
        <div key={node.id} style={{ paddingLeft: `${depth * 16 + 8}px` }}>
          <InlineNameInput
            initial={node.name}
            placeholder={isVault ? "Vault name" : "Filename"}
            onSubmit={(name) => handleRename(name, node)}
            onCancel={() => setInlineMode(null)}
            validate={isVault ? validateSpaceOrVaultName : validateLogFilename}
          />
        </div>
      );
    }

    return (
      <div key={node.id}>
        <div
          className="flex items-center gap-2 px-2 py-1.5 hover:bg-white/5 rounded-md cursor-pointer group"
          style={{ paddingLeft: `${depth * 16 + 8}px` }}
          onClick={() => {
            if (isVault) {
              toggleExpand(node.id);
            } else if (isLog) {
              onSelectLog(node.id);
            }
          }}
          onContextMenu={(e) => handleContextMenu(e, node)}
        >
          {isVault && (
            <span className="text-muted">
              {isExpanded ? <ChevronDown size={16} /> : <ChevronRight size={16} />}
            </span>
          )}
          {isLog && <span className="w-4" />}
          <span className="text-muted">
            {isVault ? <Folder size={16} /> : <File size={16} />}
          </span>
          <span className="flex-1 text-sm text-white truncate">{node.name}</span>
          <button
            onClick={(e) => {
              e.stopPropagation();
              handleContextMenu(e, node);
            }}
            className="opacity-0 group-hover:opacity-100 text-muted hover:text-white"
          >
            <MoreVertical size={14} />
          </button>
        </div>

        {/* Children and inline create inputs */}
        {isVault && isExpanded && (
          <div>
            {/* Show inline create vault input */}
            {inlineMode?.type === "create-vault" && inlineMode.parentId === node.id && (
              <div style={{ paddingLeft: `${(depth + 1) * 16 + 8}px` }}>
                <InlineNameInput
                  placeholder="Vault name"
                  onSubmit={(name) => handleCreateVault(name, node.id)}
                  onCancel={() => setInlineMode(null)}
                  validate={validateSpaceOrVaultName}
                />
              </div>
            )}

            {/* Show inline create log input */}
            {inlineMode?.type === "create-log" && inlineMode.vaultId === node.id && (
              <div style={{ paddingLeft: `${(depth + 1) * 16 + 8}px` }}>
                <InlineNameInput
                  placeholder="filename.ext"
                  onSubmit={(name) => handleCreateLog(name, node.id)}
                  onCancel={() => setInlineMode(null)}
                  validate={validateLogFilename}
                />
              </div>
            )}

            {/* Render children */}
            {node.children && node.children.length > 0 ? (
              node.children.map((child) => renderNode(child, depth + 1))
            ) : (
              !inlineMode && (
                <div className="text-sm text-muted px-2 py-2" style={{ paddingLeft: `${(depth + 1) * 16 + 8}px` }}>
                  No files yet
                </div>
              )
            )}
          </div>
        )}
      </div>
    );
  };

  const isEmpty = tree.length === 0;

  return (
    <div className="w-64 h-full bg-card border-r border-white/10 flex flex-col">
      {/* Space Selector */}
      <div className="px-4 py-3 border-b border-white/10">
        <div className="flex items-center justify-between mb-2">
          <h2 className="text-sm font-semibold text-white">Space</h2>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setInlineMode({ type: "create-space" })}
              className="text-muted hover:text-white"
              title="New Space"
              aria-label="New Space"
            >
              <Plus size={16} />
            </button>
            <button
              onClick={onOpenCommandPalette}
              className="text-muted hover:text-white"
              title="Command Palette (Ctrl+Shift+P)"
              aria-label="Command Palette"
            >
              <MoreVertical size={16} />
            </button>
          </div>
        </div>

        {/* Inline create space */}
        {inlineMode?.type === "create-space" ? (
          <InlineNameInput
            placeholder="Space name"
            onSubmit={handleCreateSpace}
            onCancel={() => setInlineMode(null)}
            validate={validateSpaceOrVaultName}
          />
        ) : (
          <select
            value={selectedSpace?.id || ""}
            onChange={(e) => {
              const space = spaces.find((s) => s.id === e.target.value);
              if (space) onSelectSpace(space);
            }}
            className="w-full bg-black border border-white/10 rounded px-2 py-1.5 text-sm text-white focus:outline-none focus:border-primaryBlue"
          >
            {spaces.map((space) => (
              <option key={space.id} value={space.id}>
                {space.name}
              </option>
            ))}
          </select>
        )}
      </div>

      {/* Tree */}
      <div className="flex-1 overflow-auto p-2">
        {isEmpty && !inlineMode ? (
          <div className="text-center py-8 px-4">
            <p className="text-sm text-muted mb-3">No Vaults yet.</p>
            <button
              onClick={() => setInlineMode({ type: "create-vault" })}
              className="text-sm text-primaryBlue hover:underline"
            >
              Create Vault
            </button>
          </div>
        ) : (
          <>
            {/* Root-level create vault input */}
            {inlineMode?.type === "create-vault" && !inlineMode.parentId && (
              <div className="px-2">
                <InlineNameInput
                  placeholder="Vault name"
                  onSubmit={(name) => handleCreateVault(name)}
                  onCancel={() => setInlineMode(null)}
                  validate={validateSpaceOrVaultName}
                />
              </div>
            )}
            {tree.map((node) => renderNode(node))}
          </>
        )}
      </div>

      {/* Context Menu */}
      {contextMenu && (
        <>
          <div className="fixed inset-0 z-40" onClick={closeContextMenu} />
          <div
            className="fixed z-50 bg-card border border-white/10 rounded-lg shadow-lg py-1 min-w-[160px]"
            style={{ top: contextMenu.y, left: contextMenu.x }}
          >
            {contextMenu.node.type === "vault" && (
              <>
                <button
                  onClick={() => {
                    setInlineMode({ type: "create-log", vaultId: contextMenu.node.id });
                    closeContextMenu();
                  }}
                  className="w-full px-4 py-2 text-left text-sm text-white hover:bg-white/5"
                >
                  New Log
                </button>
                <button
                  onClick={() => {
                    setInlineMode({ type: "create-vault", parentId: contextMenu.node.id });
                    closeContextMenu();
                  }}
                  className="w-full px-4 py-2 text-left text-sm text-white hover:bg-white/5"
                >
                  New Vault
                </button>
              </>
            )}
            <hr className="my-1 border-white/10" />
            <button
              onClick={() => {
                setInlineMode({ type: "rename", node: contextMenu.node });
                closeContextMenu();
              }}
              className="w-full px-4 py-2 text-left text-sm text-white hover:bg-white/5"
            >
              Rename
            </button>
            <button
              onClick={() => {
                handleDelete(contextMenu.node);
                closeContextMenu();
              }}
              className="w-full px-4 py-2 text-left text-sm text-accentCyan hover:bg-white/5"
            >
              Delete
            </button>
          </div>
        </>
      )}
    </div>
  );
}
