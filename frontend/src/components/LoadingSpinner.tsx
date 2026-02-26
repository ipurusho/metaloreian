export function LoadingSpinner({ message = 'Loading...' }: { message?: string }) {
  return <div className="loading-spinner">{message}</div>;
}
