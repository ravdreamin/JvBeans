"use client";

import MonacoEditor from "@monaco-editor/react";

interface EditorProps {
  language: string;
  code: string;
  onChange: (value: string) => void;
}

export default function Editor({ language, code, onChange }: EditorProps) {
  return (
    <MonacoEditor
      height="100%"
      language={language}
      value={code}
      onChange={(value) => onChange(value || "")}
      theme="vs-dark"
      options={{
        minimap: { enabled: false },
        fontSize: 14,
        lineNumbers: "on",
        scrollBeyondLastLine: false,
        automaticLayout: true,
      }}
    />
  );
}
