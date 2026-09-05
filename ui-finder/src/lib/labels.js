// Turns protobuf enum names ("RaceHuman", "ClassPaladin", "RetributionPaladin")
// into display text: strips the enum prefix, then splits camelCase words.
export function humanizeEnum(value, prefix) {
  if (!value) return '';
  const name = prefix && value.startsWith(prefix) ? value.slice(prefix.length) : value;
  return name.replace(/([a-z])([A-Z])/g, '$1 $2');
}
