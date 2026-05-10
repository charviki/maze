export function CalibrationMarks() {
  return (
    <div className="pointer-events-none absolute inset-0">
      <div className="absolute top-4 left-4 w-4 h-4 border-t border-l border-primary/30" />
      <div className="absolute top-4 right-4 w-4 h-4 border-t border-r border-primary/30" />
      <div className="absolute bottom-4 left-4 w-4 h-4 border-b border-l border-primary/30" />
      <div className="absolute bottom-4 right-4 w-4 h-4 border-b border-r border-primary/30" />

      <div className="absolute top-1/2 left-4 w-2 h-[1px] bg-primary/30" />
      <div className="absolute top-1/2 right-4 w-2 h-[1px] bg-primary/30" />
      <div className="absolute top-4 left-1/2 w-[1px] h-2 bg-primary/30" />
      <div className="absolute bottom-4 left-1/2 w-[1px] h-2 bg-primary/30" />
    </div>
  );
}
