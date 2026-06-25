export function Icon({ name, className = "" }: { name: string; className?: string }) {
  return <img className={className} src={`/assets/icons/svg/${name}.svg`} alt="" aria-hidden="true" />;
}
