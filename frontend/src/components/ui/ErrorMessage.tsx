// src/components/ui/ErrorMessage.tsx

interface ErrorMessageProps {
  message?: string;
  onRetry?: () => void;
}

export default function ErrorMessage({
  message = 'Something went wrong. Please try again.',
  onRetry,
}: ErrorMessageProps) {
  return (
    <div className="bg-red-900/20 border border-red-800/50 rounded-xl p-6 text-center animate-fade-in">
      <div className="text-4xl mb-3">⚠️</div>
      <p className="text-red-300 mb-4">{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="px-4 py-2 bg-red-800/50 hover:bg-red-800/70 text-red-200 rounded-lg transition-colors text-sm"
        >
          Try Again
        </button>
      )}
    </div>
  );
}
