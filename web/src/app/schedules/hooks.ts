import { DateTime } from 'luxon'
import { useURLParam } from '../actions'
import { getStartOfWeek } from '../util/luxon-helpers'

interface CalendarNavigation {
  weekly: boolean
  setWeekly: (val: boolean) => void
  start: string
  setStart: (val: string) => void
}

export function useCalendarNavigation(): CalendarNavigation {
  const [weekly, setWeekly] = useURLParam<boolean>('weekly', false)
  const [start, setStart] = useURLParam(
    'start',
    weekly
      ? getStartOfWeek().toISODate()
      : DateTime.now().startOf('month').toISODate(),
  )

  return {
    weekly,
    setWeekly,
    start,
    setStart,
  }
}
