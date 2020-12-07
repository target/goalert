import { DateTime } from 'luxon'
import { useMemo } from 'react'
import { useURLParam } from '../actions'

interface TimeZoneUtils {
  urlZone: string
  setUrlZone: (val: string) => void
  localZone: string
  isUrlZoneLocal: boolean
}

// useTimeZone provides time zone utilities
// urlZone: value of 'tz' URL query param, else "local"
// setUrlZone: set 'tz' URL query param
// localZone: local system time zone e.g. "America/New_York"
// isUrlZoneLocal: true if 'tz' URL query param represents local system time zone
function useTimeZone(): TimeZoneUtils {
  const [urlZone, setUrlZone] = useURLParam<string>('tz', 'local')
  const localZone = useMemo(() => DateTime.local().zone.name, [])
  const isUrlZoneLocal = urlZone === 'local' || urlZone === localZone
  return { urlZone, setUrlZone, localZone, isUrlZoneLocal }
}

export default useTimeZone
