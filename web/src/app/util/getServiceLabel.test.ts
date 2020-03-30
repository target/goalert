import getServiceLabel from './getServiceLabel'

test('it should split service labels correctly', () => {
  const check = (search: string, labelKey: string, labelValue: string): void =>
    expect(getServiceLabel(search)).toEqual({ labelKey, labelValue })

  check('wcbn.fm/rfaa=88.3 ann arbor', 'wcbn.fm/rfaa', '88.3 ann arbor')
  check('foo=bar', 'foo', 'bar')
  check('foo!=bar', 'foo', 'bar')
  check('foo=bar=baz', 'foo', 'bar=baz')
  check('foo=bar!=baz', 'foo', 'bar!=baz')
  check('foo=bar===!==', 'foo', 'bar===!==')
})
