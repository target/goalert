import getServiceFilters from './getServiceFilters'

test('it should split service labels correctly', () => {
  const check = (
    search: string,
    labelKey: string,
    labelValue: string,
    integrationKey: string,
  ): void =>
    expect(getServiceFilters(search)).toEqual({
      labelKey,
      labelValue,
      integrationKey,
    })

  check(
    'token=00000000-0000-0000-0000-000000000001 wcbn.fm/rfaa=88.3 ann arbor',
    'wcbn.fm/rfaa',
    '88.3 ann arbor',
    '00000000-0000-0000-0000-000000000001',
  )
  check(
    'token=00000000-0000-0000-0000-000000000001 foo=bar',
    'foo',
    'bar',
    '00000000-0000-0000-0000-000000000001',
  )
  check(
    'token=00000000-0000-0000-0000-000000000001 foo!=bar',
    'foo',
    'bar',
    '00000000-0000-0000-0000-000000000001',
  )
  check(
    'token=00000000-0000-0000-0000-000000000001 foo=bar=baz',
    'foo',
    'bar=baz',
    '00000000-0000-0000-0000-000000000001',
  )
  check(
    'token=00000000-0000-0000-0000-000000000001 foo=bar!=baz',
    'foo',
    'bar!=baz',
    '00000000-0000-0000-0000-000000000001',
  )
  check(
    'token=00000000-0000-0000-0000-000000000001 foo=bar===!==',
    'foo',
    'bar===!==',
    '00000000-0000-0000-0000-000000000001',
  )
})
