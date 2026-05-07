"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Bell, LayoutDashboard, Mail, FileText, Send, LogOut, ChevronsUpDown, Check, KeyRound, UserPlus, Users } from "lucide-react";
import clsx from "clsx";
import { useEffect, useRef, useState } from "react";
import { logout, listMyOrgs, selectOrg, OrgSummary } from "@/lib/auth";

const nav = [
  { href: "/dashboard", label: "Overview", icon: LayoutDashboard },
  { href: "/dashboard/providers", label: "Providers", icon: Mail },
  { href: "/dashboard/templates", label: "Templates", icon: FileText },
  { href: "/dashboard/send", label: "Send", icon: Send },
  { href: "/dashboard/deliveries", label: "Deliveries", icon: Bell },
  { href: "/dashboard/contacts", label: "Contacts", icon: Users },
  { href: "/dashboard/api-keys", label: "API Keys", icon: KeyRound },
  { href: "/dashboard/invites", label: "Invites", icon: UserPlus },
];

export default function Sidebar() {
  const pathname = usePathname();
  const router = useRouter();
  const [orgs, setOrgs] = useState<OrgSummary[]>([]);
  const [currentOrgId, setCurrentOrgId] = useState<string | null>(null);
  const [open, setOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    listMyOrgs().then(setOrgs).catch(() => {});
  }, []);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, []);

  const handleLogout = async () => {
    await logout();
    router.push("/sign-in");
  };

  const handleSelectOrg = async (orgId: string) => {
    setOpen(false);
    if (orgId === currentOrgId) return;
    try {
      await selectOrg(orgId);
      setCurrentOrgId(orgId);
      router.refresh();
    } catch {
      // ignore
    }
  };

  const currentOrg = orgs.find((o) => o.id === currentOrgId) ?? orgs[0];

  return (
    <aside className="w-56 shrink-0 flex flex-col border-r border-gray-200 bg-white min-h-screen">
      <div className="h-16 flex items-center px-4 border-b border-gray-200 bg-[#0F172A]">
        <img src="/favicons/android-chrome-512x512.png" alt="Potoo" className="h-14 w-auto" />
      </div>

      {/* Org switcher */}
      {orgs.length > 0 && (
        <div className="px-3 pt-3 pb-1 relative" ref={dropdownRef}>
          <button
            onClick={() => setOpen((v) => !v)}
            className="w-full flex items-center justify-between gap-2 px-3 py-2 rounded-md bg-gray-50 hover:bg-gray-100 transition-colors text-sm"
          >
            <span className="truncate font-medium text-gray-800">{currentOrg?.name ?? "Select org"}</span>
            <ChevronsUpDown size={14} className="text-gray-400 shrink-0" />
          </button>
          {open && (
            <div className="absolute left-3 right-3 top-full mt-1 bg-white border border-gray-200 rounded-md shadow-lg z-50 py-1">
              {orgs.map((org) => (
                <button
                  key={org.id}
                  onClick={() => handleSelectOrg(org.id)}
                  className="w-full flex items-center justify-between gap-2 px-3 py-2 text-sm hover:bg-gray-50 transition-colors"
                >
                  <div className="text-left truncate">
                    <p className="font-medium text-gray-800 truncate">{org.name}</p>
                    <p className="text-xs text-gray-400">{org.role}</p>
                  </div>
                  {currentOrg?.id === org.id && <Check size={13} className="text-gray-500 shrink-0" />}
                </button>
              ))}
            </div>
          )}
        </div>
      )}

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
