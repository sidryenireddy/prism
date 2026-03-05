"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import Image from "next/image";

const navItems = [
  { href: "/", label: "Analyses" },
  { href: "/dashboard", label: "Dashboards" },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-56 bg-black text-white flex flex-col h-screen shrink-0">
      <div className="p-4 border-b border-neutral-800">
        <Link href="/" className="flex items-center gap-2">
          <Image
            src="/rebelicon.png"
            alt="Prism"
            width={28}
            height={28}
            className="rounded"
          />
          <span className="text-lg font-semibold tracking-tight">Prism</span>
        </Link>
      </div>
      <nav className="flex-1 p-3 space-y-1">
        {navItems.map((item) => {
          const active =
            item.href === "/"
              ? pathname === "/"
              : pathname.startsWith(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`block px-3 py-2 rounded text-sm transition-colors ${
                active
                  ? "bg-neutral-800 text-white"
                  : "text-neutral-400 hover:text-white hover:bg-neutral-900"
              }`}
            >
              {item.label}
            </Link>
          );
        })}
      </nav>
      <div className="p-4 border-t border-neutral-800 text-xs text-neutral-500">
        Prism v1.0
      </div>
    </aside>
  );
}
