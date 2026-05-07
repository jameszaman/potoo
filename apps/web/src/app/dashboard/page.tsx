export default function DashboardPage() {
  return (
    <div>
      <h1 className="text-xl font-semibold mb-1">Overview</h1>
      <p className="text-sm text-gray-500 mb-8">Connect a provider, create a template, then send notifications via API.</p>

      <div className="grid grid-cols-2 gap-4">
        {[
          { label: "Providers", desc: "Connect Resend or SendGrid to send emails", href: "/dashboard/providers" },
          { label: "Templates", desc: "Create and manage email templates with variables", href: "/dashboard/templates" },
          { label: "Send", desc: "Send a test notification from the dashboard", href: "/dashboard/send" },
          { label: "Deliveries", desc: "View delivery logs, status, and event timelines", href: "/dashboard/deliveries" },
          { label: "Contacts", desc: "Save recipients and group them by tag for easy reuse", href: "/dashboard/contacts" },
          { label: "API Keys", desc: "Generate keys to authenticate API requests from your app", href: "/dashboard/api-keys" },
          { label: "Invites", desc: "Create single-use invite links for your organization", href: "/dashboard/invites" },
        ].map((card) => (
          <a
            key={card.href}
            href={card.href}
            className="block bg-white border border-gray-200 rounded-lg p-5 hover:border-gray-400 transition-colors shadow-sm"
          >
            <p className="font-medium text-sm mb-0.5">{card.label}</p>
            <p className="text-xs text-gray-500">{card.desc}</p>
          </a>
        ))}
      </div>
    </div>
  );
}
