import { cn, clipPathHalf } from '../../utils';

interface SkeletonProps {
  className?: string;
}

export function Skeleton({ className }: SkeletonProps) {
  return (
    <div
      className={cn('animate-pulse bg-primary/10', className)}
      style={{
        clipPath: clipPathHalf(6),
      }}
    />
  );
}
