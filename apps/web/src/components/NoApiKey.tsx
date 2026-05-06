"use client";

import Link from "next/link";
import { KeyRound } from "lucide-react";

export function isApiKeyError(error: string): boolean {
  const msg = error.toLowerCase();
  return (
    msg.includes("authorization header is required") ||
    msg.includes("api key is invalid") ||
    msg.includes("missing_api_key") ||
    msg.includes("invalid_api_key")
  );
}

export default function NoApiKey() {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center">
      <div className="w-10 h-10 rounded-full bg-gray-100 flex items-center justify-center mb-4">
        <KeyRound size={18} className="text-gray-500" />
      </div>
      <h2 className="text-sm font-semibold text-gray-900 mb-1">No API key configured</h2>
      <p className="text-sm text-gray-500 mb-4 max-w-xs">
        This section requires an API key. Create one in your settings and add it to your requests.
      </p>
      <Link
        href="/dashboard/api-keys"
        className="text-sm font-medium text-gray-900 underline underline-offset-2 hover:text-gray-600 transition-colors"
      >
        Go to API keys →
      </Link>
    </div>
  );
}
