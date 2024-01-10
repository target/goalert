import { DateTime } from 'luxon'
import { useURLParams } from '../../actions/hooks'
import { getStartOfWeek } from '../luxon-helpers'

interface CalendarNavParams {
  weekly: boolean
  start: string
}

interface CalendarNavigation extends CalendarNavParams {
  setParams: (val: Partial<CalendarNavParams>) => void
}

export function useCalendarNavigation(): CalendarNavigation {
  const [_params] = useURLParams({ weekly: false })
  const [params, setParams] = useURLParams({
    weekly: false as boolean,
    start: _params.weekly
      ? getStartOfWeek().toISODate()
      : DateTime.now().startOf('month').toISODate(),
  })

  return {
    weekly: params.weekly,
    start: params.start,
    setParams,
  }
}
