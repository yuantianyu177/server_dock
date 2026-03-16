export function formatIPv4Ports(ports) {
  if (!ports) return '-'

  const entries = String(ports)
    .split(',')
    .map((entry) => entry.trim())
    .filter(Boolean)

  const ipv4Entries = entries.filter((entry) => !entry.startsWith('[::]:'))
  return (ipv4Entries.length > 0 ? ipv4Entries : entries).join(', ')
}
