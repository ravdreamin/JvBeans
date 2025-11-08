"use client";

import { useState } from "react";
import { ChevronRight, ChevronDown, Folder, File, Plus, MoreVertical } from "lucide-react";
import { createSpace, createVault, createLog, updateVault, deleteVault, updateLog, deleteLog as deleteLogAPI } from "@/lib/api";
import type { Space, TreeNode } from "@/lib/types";

interface ExplorerSidebarProps {
  spaces: Space[];
  selectedSpace: Space | null;
  tree: TreeNode[];
  onSelectSpace: (space: Space) => void;
  onSelectLog: (logId: string) => void;
  onRefresh: () => void;
  onSpacesChange: () => void;
}

export default function ExplorerSidebar({
  spaces,
  selectedSpace,
  tree,
  onSelectSpace,
  onSelectLog,
  onRefresh,
  onSpacesChange,
}: ExplorerSidebarProps) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [contextMenu, setContextMenu] = useState<{ x: number; y: number; node: TreeNode | null } | null>(null);

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

  const handleContextMenu = (e: React.MouseEvent, node: TreeNode) => {
    e.preventDefault();
    e.stopPropagation();
    setContextMenu({ x: e.clientX, y: e.clientY, node });
  };

  const closeContextMenu = () => {
    setContextMenu(null);
  };

  const handleNewSpace = async () => {
    const name = prompt("Enter space name:");
    if (!name) return;

    try {
      await createSpace(name);
      onSpacesChange();
    } catch (error) {
      alert("Failed to create space");
    }
  };

  const handleNewVault = async (parentNode?: TreeNode) => {
    if (!selectedSpace) return;

    const name = prompt("Enter vault name:");
    if (!name) return;

    try {
      const parentId = parentNode?.type === "vault" ? parentNode.id : undefined;
      await createVault(selectedSpace.id, name, parentId);
      onRefresh();
    } catch (error) {
      alert("Failed to create vault");
    }
    closeContextMenu();
  };

  const handleNewLog = async (vaultNode?: TreeNode) => {
    if (!selectedSpace) return;

    const filename = prompt("Enter filename (e.g., app.js, main.py):");
    if (!filename) return;

    try {
      const vaultId = vaultNode?.type === "vault" ? vaultNode.id : undefined;
      if (!vaultId) {
        alert("Please create a vault first");
        return;
      }
      await createLog(selectedSpace.id, vaultId, filename);
      onRefresh();
    } catch (error) {
      alert("Failed to create log");
    }
    closeContextMenu();
  };

  const handleRename = async () => {
    if (!contextMenu?.node) return;

    const node = contextMenu.node;
    const newNameValue = prompt(`Rename ${node.name}:`, node.name);
    if (!newNameValue || newNameValue === node.name) {
      closeContextMenu();
      return;
    }

    try {
      if (node.type === "vault") {
        await updateVault(node.id, newNameValue);
      } else if (node.type === "log") {
        await updateLog(node.id, { name: newNameValue });
      }
      onRefresh();
    } catch (error) {
      alert("Failed to rename");
    }
    closeContextMenu();
  };

  const handleDelete = async () => {
    if (!contextMenu?.node) return;

    const node = contextMenu.node;
    if (!confirm(`Delete ${node.name}?`)) {
      closeContextMenu();
      return;
    }

    try {
      if (node.type === "vault") {
        await deleteVault(node.id);
      } else if (node.type === "log") {
        await deleteLogAPI(node.id);
      }
      onRefresh();
    } catch (error) {
      alert("Failed to delete");
    }
    closeContextMenu();
  };

  const handleDuplicate = async () => {
    if (!contextMenu?.node || contextMenu.node.type !== "log") return;
    // Not implemented in backend yet
    alert("Duplicate feature coming soon");
    closeContextMenu();
  };

  const renderNode = (node: TreeNode, depth: number = 0) => {
    const isExpanded = expanded.has(node.id);
    const isVault = node.type === "vault";
    const isLog = node.type === "log";

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
        {isVault && isExpanded && node.children && (
          <div>
            {node.children.length === 0 ? (
              <div className="text-sm text-muted px-2 py-2" style={{ paddingLeft: `${(depth + 1) * 16 + 8}px` }}>
                This Vault has no Logs.
              </div>
            ) : (
              node.children.map((child) => renderNode(child, depth + 1))
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
          <button
            onClick={handleNewSpace}
            className="text-muted hover:text-white"
            title="New Space"
            aria-label="New Space"
          >
            <Plus size={16} />
          </button>
        </div>
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
      </div>

      {/* Tree */}
      <div className="flex-1 overflow-auto p-2">
        {isEmpty ? (
          <div className="text-center py-8 px-4">
            <p className="text-sm text-muted mb-3">No Vaults yet.</p>
            <button
              onClick={() => handleNewVault()}
              className="text-sm text-primaryBlue hover:underline"
            >
              Create Vault
            </button>
          </div>
        ) : (
          tree.map((node) => renderNode(node))
        )}
      </div>

      {/* Context Menu */}
      {contextMenu && (
        <>
          <div
            className="fixed inset-0 z-40"
            onClick={closeContextMenu}
          />
          <div
            className="fixed z-50 bg-card border border-white/10 rounded-lg shadow-lg py-1 min-w-[160px]"
            style={{ top: contextMenu.y, left: contextMenu.x }}
          >
            {contextMenu.node?.type === "vault" && (
              <>
                <button
                  onClick={() => handleNewLog(contextMenu.node!)}
                  className="w-full px-4 py-2 text-left text-sm text-white hover:bg-white/5"
                >
                  New Log
                </button>
                <button
                  onClick={() => handleNewVault(contextMenu.node!)}
                  className="w-full px-4 py-2 text-left text-sm text-white hover:bg-white/5"
                >
                  New Vault
                </button>
              </>
            )}
            {contextMenu.node?.type === "log" && (
              <button
                onClick={handleDuplicate}
                className="w-full px-4 py-2 text-left text-sm text-white hover:bg-white/5"
              >
                Duplicate
              </button>
            )}
            <hr className="my-1 border-white/10" />
            <button
              onClick={handleRename}
              className="w-full px-4 py-2 text-left text-sm text-white hover:bg-white/5"
            >
              Rename
            </button>
            <button
              onClick={handleDelete}
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
