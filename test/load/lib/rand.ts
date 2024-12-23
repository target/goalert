export function randString(
  len: number = 32,
  chars: string = 'abcdefghijklmnopqrstuvwxyz',
): string {
  let str = ''
  for (let i = 0; i < len; i++) {
    str += chars[Math.floor(Math.random() * chars.length)]
  }
  return str
}

export function randEmail(): string {
  return `${randString(16)}@${randString(12)}.com`
}

export function randInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1) + min)
}

export function randBool(): boolean {
  return Math.random() < 0.5
}

export function randDate(): Date {
  return new Date(randInt(0, Date.now()))
}

export function randTime(): string {
  return `${randInt(0, 23).toString().padStart(2, '0')}:${randInt(0, 59).toString().padStart(2, '0')}`
}

const timeZones = [
  'Etc/UTC',
  'America/Los_Angeles',
  'America/New_York',
  'America/Chicago',
  'Europe/London',
  'Asia/Tokyo',
  'Australia/Sydney',
]

export function randPickOne<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)]
}

export function randTimeZone(): string {
  return randPickOne(timeZones)
}

/* randSample returns a random sample of items from the input array (duplicates possible). */
export function randSample<T>(
  items: T[],
  max: number = items.length,
  min: number = 0,
): T[] {
  const n = randInt(min, max)
  const result = []
  for (let i = 0; i < n; i++) {
    result.push(randPickOne(items))
  }
  return result
}

/* randSampleUnique returns a random sample of unique items from the input array. */
export function randSampleUnique<T>(
  items: T[],
  min: number = 0,
  max: number = items.length,
): T[] {
  if (max > items.length) {
    max = items.length
  }
  const n = randInt(min, max)
  const result: T[] = []
  for (let i = 0; i < n; i++) {
    const item = randPickOne(items)
    if (!result.includes(item)) {
      result.push(item)
    }
  }
  return result
}
