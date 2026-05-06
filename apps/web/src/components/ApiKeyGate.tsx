"use client";

import { useState, useEffect } from "react";

const STORAGE_KEY = "nl_api_key";

export function useApiKey() {
  const [apiKey, setApiKeyState] = useState<string>("");

  useEffect(() => {
    setApiKeyState(localStorage.getItem(STORAGE_KEY) ?? "");
  }, []);

  const setApiKey = (key: string) => {
    localStorage.setItem(STORAGE_KEY, key);
    setApiKeyState(key);
  };

  return { apiKey, setApiKey };
}

export default function ApiKeyGate({ children }: { children: React.ReactNode }) {
  const { apiKey, setApiKey } = useApiKey();
  const [input, setInput] = useState("");

  if (!apiKey) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="bg-white border border-gray-200 rounded-lg p-8 w-full max-w-sm shadow-sm">
          <h2 className="text-base font-semibold mb-1">Enter your API key</h2>
          <p className="text-sm text-gray-500 mb-4">Your key is stored only in this browser.</p>
          <input
            type="password"
            placeholder="sk_live_..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && input && setApiKey(input)}
            className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm mb-3 focus:outline-none focus:ring-2 focus:ring-gray-900"
          />
          <button
            onClick={() => input && setApiKey(input)}
            className="w-full bg-gray-900 text-white text-sm font-medium py-2 rounded-md hover:bg-gray-700 transition-colors"
          >
            Save
          </button>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
