export default function DashboardPage() {
  return (
    <div>
      <h1 className="text-xl font-semibold mb-1">Overview</h1>
      <p className="text-sm text-gray-500 mb-8">Welcome to NotifyLayer.</p>

      <div className="grid grid-cols-2 gap-4">
        {[
          { label: "Providers", desc: "Connect Resend or SendGrid", href: "/dashboard/providers" },
          { label: "Templates", desc: "Create and manage email templates", href: "/dashboard/templates" },
          { label: "Send", desc: "Send a test notification", href: "/dashboard/send" },
          { label: "Deliveries", desc: "View delivery logs and timelines", href: "/dashboard/deliveries" },
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
