"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function EditorPage() {
  const router = useRouter();

  useEffect(() => {
    router.replace("/flow");
  }, [router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-black text-white">
      <p>Redirecting to Flow...</p>
    </div>
  );
}
