export async function runSettledBatch(items, task, concurrency = 32) {
  const results = new Array(items.length)
  let nextIndex = 0

  async function worker() {
    while (nextIndex < items.length) {
      const index = nextIndex
      nextIndex += 1
      try {
        results[index] = { status: 'fulfilled', value: await task(items[index], index) }
      } catch (reason) {
        results[index] = { status: 'rejected', reason }
      }
    }
  }

  const workerCount = Math.min(Math.max(1, concurrency), items.length)
  await Promise.all(Array.from({ length: workerCount }, worker))
  return results
}

export function summarizeBatchResults(items, results) {
  const failedItems = items.filter((_, index) => results[index].status === 'rejected')
  const firstFailure = results.find(result => result.status === 'rejected')

  return {
    failedItems,
    succeededCount: items.length - failedItems.length,
    firstError: firstFailure?.reason?.message
  }
}
