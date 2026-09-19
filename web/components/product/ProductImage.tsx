import Image from "next/image";

export function ProductImage({
  src,
  alt,
  width,
  height,
  priority = false,
  className = "size-full object-contain",
}: {
  src?: string;
  alt: string;
  width: number;
  height: number;
  priority?: boolean;
  className?: string;
}) {
  if (!src) {
    return <span className="text-xs text-muted">بدون تصویر</span>;
  }
  return (
    <Image
      src={src}
      alt={alt}
      width={width}
      height={height}
      priority={priority}
      className={className}
    />
  );
}
