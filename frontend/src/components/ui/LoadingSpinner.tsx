// src/components/ui/LoadingSpinner.tsx

export default function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center py-12">
      <div className="relative">
        <div className="w-12 h-12 rounded-full border-4 border-islamic-900/30"></div>
        <div className="w-12 h-12 rounded-full border-4 border-islamic-500 border-t-transparent animate-spin absolute top-0 left-0"></div>
      </div>
    </div>
  );
}
