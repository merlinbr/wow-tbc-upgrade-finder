export function label(value) {
  if (!value) {
    return '';
  }
  return String(value)
    .split('_')
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ');
}
