"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Bell, LayoutDashboard, Mail, FileText, Send, LogOut } from "lucide-react";
import clsx from "clsx";
import { logout } from "@/lib/auth";

const nav = [
  { href: "/dashboard", label: "Overview", icon: LayoutDashboard },
  { href: "/dashboard/providers", label: "Providers", icon: Mail },
  { href: "/dashboard/templates", label: "Templates", icon: FileText },
  { href: "/dashboard/send", label: "Send", icon: Send },
  { href: "/dashboard/deliveries", label: "Deliveries", icon: Bell },
];

export default function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();

  const handleLogout = async () => {
    await logout();
    router.push("/sign-in");
  };

  return (
    <aside className="w-56 shrink-0 flex flex-col border-r border-gray-200 bg-white min-h-screen">
      <div className="h-14 flex items-center px-5 border-b border-gray-200">
        <span className="font-semibold text-gray-900 tracking-tight">NotifyLayer</span>
      </div>

      <nav className="flex-1 py-4 px-3 space-y-0.5">
        {nav.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || (href !== "/dashboard" && pathname.startsWith(href));
          return (
            <Link
              key={href}
              href={href}
              className={clsx(
                "flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium transition-colors",
                active
                  ? "bg-gray-100 text-gray-900"
                  : "text-gray-500 hover:text-gray-900 hover:bg-gray-50"
              )}
            >
              <Icon size={16} />
              {label}
            </Link>
          );
        })}
      </nav>

      <div className="p-4 border-t border-gray-200">
        <button
          onClick={handleLogout}
          className="flex items-center gap-2 text-sm text-gray-500 hover:text-gray-900 transition-colors w-full"
        >
          <LogOut size={15} />
          Sign out
        </button>
      </div>
    </aside>
  );
}
