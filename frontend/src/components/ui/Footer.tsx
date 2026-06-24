// src/components/ui/Footer.tsx

export default function Footer() {
  return (
    <footer className="border-t border-dark-border py-6 mt-8">
      <div className="max-w-4xl mx-auto px-4 text-center">
        <p className="text-dark-textMuted text-sm">
          © {new Date().getFullYear()} Navisha Prayer · Made with ☪
        </p>
        <p className="text-dark-textMuted/50 text-xs mt-1">
          Powered by Aladhan API
        </p>
      </div>
    </footer>
  );
}
