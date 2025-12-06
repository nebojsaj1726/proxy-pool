type Props = React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "outline";
  textColor?: string;
};

export function Button({
  variant = "primary",
  className = "",
  textColor,
  ...props
}: Props) {
  const base = "px-3 py-1.5 rounded text-sm font-medium transition";
  const variants = {
    primary: "bg-blue-500 text-white hover:bg-blue-700",
    outline: `border border-gray-300 hover:bg-gray-700 ${
      textColor || "text-gray-300"
    }`,
  };
  return (
    <button
      {...props}
      className={`${base} ${variants[variant]} ${className}`}
    />
  );
}
