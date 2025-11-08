"use client";

import Link from "next/link";
import dynamic from "next/dynamic";
import { useEffect, useState } from "react";

// Spline: client-only
const Spline = dynamic(
  () => import("@splinetool/react-spline").then(m => m.default),
  { ssr: false }
);

// Your URL as PRIMARY
const PRIMARY_SCENE =
  "https://prod.spline.design/0X7gdXtZncePk4iD/scene.splinecode";
const SECONDARY_SCENE =
  "https://prod.spline.design/Bo31s7nbQkYxKFCi/scene.splinecode";

/* -------------------- Spline background -------------------- */
function SplineBackground() {
  const [sceneUrl, setSceneUrl] = useState(PRIMARY_SCENE);
  const [loaded, setLoaded] = useState(false);
  const [failed, setFailed] = useState(false);
  const [reduced, setReduced] = useState(false); // read on client only

  useEffect(() => {
    if (typeof window !== "undefined") {
      setReduced(window.matchMedia("(prefers-reduced-motion: reduce)").matches);
    }
  }, []);

  // one retry after 6s if primary never loads
  useEffect(() => {
    const t = setTimeout(() => {
      if (!loaded && sceneUrl === PRIMARY_SCENE) setSceneUrl(SECONDARY_SCENE);
    }, 6000);
    return () => clearTimeout(t);
  }, [loaded, sceneUrl]);

  if (failed) {
    return (
      <div className="absolute inset-0 -z-10 bg-gradient-to-br from-primaryBlue/30 via-black to-primaryTeal/30" />
    );
  }

  return (
    <div
      className={`absolute inset-0 -z-10 pointer-events-none ${
        reduced ? "opacity-75" : ""
      }`}
    >
      <div className="w-full h-full">
        <Spline
          key={sceneUrl}
          scene={sceneUrl}
          onLoad={() => setLoaded(true)}
          onError={() => {
            if (sceneUrl === PRIMARY_SCENE) setSceneUrl(SECONDARY_SCENE);
            else setFailed(true);
          }}
        />
      </div>
    </div>
  );
}

/* -------------------- Page -------------------- */
export default function Home() {
  const scrollTo = (id: string) =>
    document.getElementById(id)?.scrollIntoView({ behavior: "smooth" });

  return (
    <div className="overflow-x-hidden bg-black text-white">
      {/* HERO */}
      <section className="relative min-h-screen flex flex-col items-center justify-center text-center">
        <SplineBackground />
        <div className="absolute inset-0 -z-0 bg-gradient-to-b from-black/80 via-black/70 to-black/90" />

        <div className="relative z-10 px-4 max-w-4xl mx-auto">
          <h1 className="text-6xl md:text-7xl font-extrabold mb-6">
            Code, Generate,{" "}
            <span className="bg-gradient-to-r from-primaryBlue via-accentCyan to-primaryTeal bg-clip-text text-transparent">
              Execute.
            </span>
          </h1>
          <p className="text-xl md:text-2xl text-gray-300 mb-10 max-w-2xl mx-auto">
            A browser-based coding platform with AI-powered generation. Write,
            run, and save your code across languages in real time.
          </p>
          <div className="flex justify-center gap-6">
            <Link
              href="/flow"
              className="px-10 py-4 text-lg font-bold bg-primaryBlue text-white rounded-full hover:shadow-lg hover:shadow-primaryBlue/40 transition"
            >
              Try Flow
            </Link>
            <button
              onClick={() => scrollTo("features")}
              className="px-10 py-4 text-lg font-bold border border-white/30 rounded-full hover:bg-white/10 transition"
            >
              Learn More
            </button>
          </div>
        </div>
      </section>

      {/* FEATURES */}
      <section id="features" className="py-32 bg-black">
        <div className="max-w-6xl mx-auto px-4">
          <h2 className="text-5xl font-bold text-center mb-16">Features</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {[
              ["🌐", "Multi-Language Support", "Run Python, JS, Java, C, C++, Go via Piston API."],
              ["✨", "AI Code Generation", "OpenAI primary with Gemini fallback for snippets."],
              ["📁", "Organized Workspace", "Spaces, Vaults, and Logs hierarchy."],
              ["⚡", "Instant Execution", "Stdout, stderr, and exit codes in real time."],
              ["🎨", "Monaco Editor", "Syntax highlighting and shortcuts."],
              ["🔒", "Admin Mode", "Single-tenant workspace with optional token."],
            ].map(([icon, title, text], i) => (
              <div
                key={i}
                className="bg-card border border-white/10 rounded-2xl p-8 hover:border-primaryBlue/50 transition-all"
              >
                <div className="text-4xl mb-4">{icon}</div>
                <h3 className="text-2xl font-bold mb-3">{title}</h3>
                <p className="text-gray-400 leading-relaxed">{text}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* FOOTER */}
      <footer className="border-t border-white/10 py-8 bg-black">
        <div className="max-w-7xl mx-auto px-4 flex justify-between items-center text-gray-400 text-sm">
          <a
            href={process.env.NEXT_PUBLIC_GITHUB_URL || "https://github.com/ravdreamin"}
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-white"
          >
            GitHub
          </a>
          <p>© {new Date().getFullYear()} CodeFlow</p>
        </div>
      </footer>
    </div>
  );
}
