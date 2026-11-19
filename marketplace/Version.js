function fromOutput(output) {
  const match = String(output).match(/\bomarchy-files\s+v?([0-9]+\.[0-9]+\.[0-9]+)/i)
  return match ? match[1] : ""
}

function compare(left, right) {
  const a = String(left).split(".").map(Number)
  const b = String(right).split(".").map(Number)
  if (a.length < 3 || b.length < 3 || a.some(isNaN) || b.some(isNaN)) return 0
  for (let i = 0; i < 3; i++) {
    if (a[i] !== b[i]) return a[i] < b[i] ? -1 : 1
  }
  return 0
}
