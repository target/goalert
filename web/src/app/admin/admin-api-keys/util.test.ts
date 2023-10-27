import { selectedDaysUntilTimestamp } from './util'

describe('selectedDaysUntilTimestamp', () => {
  it('returns 0 if there are no matching options', () => {
    expect(
      selectedDaysUntilTimestamp(
        '2020-01-01T00:00:00Z',
        [1, 2, 3],
        '2020-01-01T00:00:00Z',
      ),
    ).toEqual(0)

    expect(
      selectedDaysUntilTimestamp(
        '2020-01-09T00:00:00Z',
        [1, 2, 3],
        '2020-01-01T00:00:00Z',
      ),
    ).toEqual(0)
  })

  it('returns the number of days since the timestamp', () => {
    expect(
      selectedDaysUntilTimestamp(
        '2020-01-02T00:00:00Z',
        [1, 2, 3],
        '2020-01-01T00:00:00Z',
      ),
    ).toEqual(1)
    expect(
      selectedDaysUntilTimestamp(
        '2020-01-04T00:00:00Z',
        [1, 2, 3],
        '2020-01-01T00:00:00Z',
      ),
    ).toEqual(3)
  })

  it('returns the number of days since the timestamp, even if there are slight differences in the time', () => {
    expect(
      selectedDaysUntilTimestamp(
        '2020-01-02T01:00:00Z',
        [1, 2, 3],
        '2020-01-01T00:00:01Z',
      ),
    ).toEqual(1)

    expect(
      selectedDaysUntilTimestamp(
        '2020-01-04T01:00:01Z',
        [1, 2, 3],
        '2020-01-01T00:00:00Z',
      ),
    ).toEqual(3)

    expect(
      selectedDaysUntilTimestamp(
        '2023-10-12T17:52:21.467Z',
        [7, 14, 30, 60, 90],
        '2023-10-05T17:52:23.782Z',
      ),
    ).toEqual(7)
  })
})
