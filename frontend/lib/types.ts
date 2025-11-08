export interface Space {
  id: string;
  userId: string;
  name: string;
  createdAt: string;
  updatedAt: string;
}

export interface Vault {
  id: string;
  spaceId: string;
  userId: string;
  name: string;
  path: string;
  parentId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Log {
  id: string;
  spaceId: string;
  vaultId: string;
  userId: string;
  name: string;
  path: string;
  language: string;
  code: string;
  createdAt: string;
  updatedAt: string;
}

export interface TreeNode {
  id: string;
  name: string;
  type: "space" | "vault" | "log";
  language?: string;
  path: string;
  children?: TreeNode[];
}

export interface RunResult {
  stdout: string;
  stderr: string;
  code: number;
  output: string;
}

export interface GenerateResponse {
  code: string;
  provider: string;
}
