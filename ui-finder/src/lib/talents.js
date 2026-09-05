// Splits a Wowhead talent string into one rank string per tree.
// Trees are hyphen-separated; each character is one talent's rank, in the
// array order used by ui/core/talents/trees/*.json. Missing trees become ''.
export function decodeTalentsString(talentsString, treeCount = 3) {
  const parts = String(talentsString ?? '').split('-');
  return Array.from({ length: treeCount }, (_, index) => parts[index] ?? '');
}

// Rank allocated at one talent index; missing digits count as zero.
export function rankAt(treeRanks, index) {
  const value = Number(String(treeRanks ?? '').charAt(index));
  return Number.isInteger(value) && value > 0 ? value : 0;
}

// Total points allocated in one tree's rank string.
export function treePoints(treeRanks) {
  let sum = 0;
  for (const digit of String(treeRanks ?? '')) {
    const value = Number(digit);
    if (Number.isInteger(value)) sum += value;
  }
  return sum;
}
