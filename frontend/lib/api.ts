import axios from "axios";
import type { Space, Vault, Log, TreeNode, RunResult, GenerateResponse } from "./types";

// Use environment variables or fallback
const getApiUrl = () => {
  if (typeof window !== "undefined") {
    return process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
  }
  return "http://localhost:8080";
};

const getAdminToken = () => {
  if (typeof window !== "undefined") {
    return process.env.NEXT_PUBLIC_ADMIN_TOKEN || "";
  }
  return "";
};

const api = axios.create({
  baseURL: getApiUrl(),
  headers: {
    "Content-Type": "application/json",
  },
});

// Attach admin token on write operations
api.interceptors.request.use((config) => {
  const token = getAdminToken();
  if (token && config.method && ["post", "put", "patch", "delete"].includes(config.method.toLowerCase())) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Spaces
export const getSpaces = async (): Promise<Space[]> => {
  const { data } = await api.get("/api/spaces");
  return data;
};

export const getSpace = async (id: string): Promise<Space> => {
  const { data } = await api.get(`/api/spaces/${id}`);
  return data;
};

export const createSpace = async (name: string): Promise<Space> => {
  const { data } = await api.post("/api/spaces", { name });
  return data;
};

export const updateSpace = async (id: string, name: string): Promise<Space> => {
  const { data } = await api.put(`/api/spaces/${id}`, { name });
  return data;
};

export const deleteSpace = async (id: string): Promise<void> => {
  await api.delete(`/api/spaces/${id}`);
};

// Tree
export const getTree = async (spaceId: string): Promise<TreeNode[]> => {
  const { data } = await api.get(`/api/tree`, { params: { spaceId } });
  return data;
};

// Vaults
export const getVaults = async (spaceId: string): Promise<Vault[]> => {
  const { data } = await api.get("/api/vaults", { params: { spaceId } });
  return data;
};

export const createVault = async (spaceId: string, name: string, parentId?: string): Promise<Vault> => {
  const { data } = await api.post("/api/vaults", { spaceId, name, parentId });
  return data;
};

export const updateVault = async (id: string, name: string): Promise<Vault> => {
  const { data } = await api.put(`/api/vaults/${id}`, { name });
  return data;
};

export const deleteVault = async (id: string): Promise<void> => {
  await api.delete(`/api/vaults/${id}`);
};

// Logs
export const getLogs = async (spaceId?: string, vaultId?: string): Promise<Log[]> => {
  const { data } = await api.get("/api/logs", { params: { spaceId, vaultId } });
  return data;
};

export const getLog = async (id: string): Promise<Log> => {
  const { data } = await api.get(`/api/logs/${id}`);
  return data;
};

export const createLog = async (spaceId: string, vaultId: string, name: string, code?: string): Promise<Log> => {
  const { data } = await api.post("/api/logs", { spaceId, vaultId, name, code: code || "" });
  return data;
};

export const updateLog = async (id: string, updates: { name?: string; code?: string }): Promise<Log> => {
  const { data } = await api.put(`/api/logs/${id}`, updates);
  return data;
};

export const deleteLog = async (id: string): Promise<void> => {
  await api.delete(`/api/logs/${id}`);
};

// Run
export const runCode = async (language: string, code: string): Promise<RunResult> => {
  const { data } = await api.post("/api/run", { language, code });
  return data;
};

// AI Generate
export const generateCode = async (prompt: string, language?: string): Promise<GenerateResponse> => {
  const { data } = await api.post("/api/ai/generate", { prompt, language });
  return data;
};
