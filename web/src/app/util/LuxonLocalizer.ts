import {
  DateLocalizer,
  DateRange,
  DateRangeFormatFunction,
  Formats,
} from 'react-big-calendar'
import { DateTime } from 'luxon'

const isSameMonth = (range: DateRange): boolean => {
  return (
    range.start.getFullYear() === range.end.getFullYear() &&
    range.start.getMonth() === range.end.getMonth()
  )
}

const dateRangeFormat: DateRangeFormatFunction = (
  range,
  culture,
  localizer,
): string =>
  `${localizer?.format(range.start, 'D', culture)} — ${localizer?.format(range.end, 'D', culture)}`

const timeRangeFormat: DateRangeFormatFunction = (
  range,
  culture,
  localizer,
): string =>
  `${localizer?.format(range.start, 't', culture)} — ${localizer?.format(range.end, 't', culture)}`

const timeRangeStartFormat: DateRangeFormatFunction = (
  range,
  culture,
  localizer,
): string => `${localizer?.format(range.start, 't', culture)} — `

const timeRangeEndFormat: DateRangeFormatFunction = (
  range,
  culture,
  localizer,
): string => ` — ${localizer?.format(range.end, 't', culture)}`

const weekRangeFormat: DateRangeFormatFunction = (
  range,
  culture,
  localizer,
): string =>
  `${localizer?.format(range.start, 'MMMM dd', culture)} — ${localizer?.format(
    range.end,
    isSameMonth(range) ? 'dd' : 'MMMM dd',
    culture,
  )}`

export const formats: Formats = {
  dateFormat: 'dd',
  dayFormat: 'dd EEE',
  weekdayFormat: 'ccc',

  selectRangeFormat: timeRangeFormat,
  eventTimeRangeFormat: timeRangeFormat,
  eventTimeRangeStartFormat: timeRangeStartFormat,
  eventTimeRangeEndFormat: timeRangeEndFormat,

  timeGutterFormat: 't',

  monthHeaderFormat: 'MMMM yyyy',
  dayHeaderFormat: 'cccc MMM dd',
  dayRangeHeaderFormat: weekRangeFormat,
  agendaHeaderFormat: dateRangeFormat,

  agendaDateFormat: 'ccc MMM dd',
  agendaTimeFormat: 't',
  agendaTimeRangeFormat: timeRangeFormat,
}

interface LuxonLocalizerOptions {
  firstDayOfWeek: number
}

const LuxonLocalizer = (
  DateTime: typeof import('luxon').DateTime,
  { firstDayOfWeek }: LuxonLocalizerOptions,
): DateLocalizer => {
  const locale = (d: DateTime, c?: string): DateTime =>
    c ? d.reconfigure({ locale: c }) : d

  return new DateLocalizer({
    formats,
    firstOfWeek() {
      return firstDayOfWeek
    },

    format(value, format, culture) {
      return locale(DateTime.fromJSDate(value as Date), culture).toFormat(
        format,
      )
    },
  })
}

export default LuxonLocalizer
