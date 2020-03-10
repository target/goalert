import {
  urlParamSelector,
  searchSelector,
  absURLSelector,
  ReduxState,
} from './url'

describe('urlParamSelector', () => {
  ;([
    { search: '?search=foo', expected: { search: 'foo' } },
    { search: '', expected: {} },
    { search: '?&search=foo', expected: { search: 'foo' } },
    { search: '?search=foo&&&', expected: { search: 'foo' } },
    { search: '?foo=bar&bin=baz', expected: { foo: 'bar', bin: 'baz' } },
    { search: '?search=asdf%26%3D', expected: { search: 'asdf&=' } },
  ] as { search: string; expected: { [index: string]: string } }[]).forEach(
    cfg =>
      test(cfg.search || '(empty)', () => {
        const res = urlParamSelector({
          router: {
            location: { search: cfg.search },
          },
        } as ReduxState)
        for (const key in cfg.expected) {
          expect(res(key)).toBe(cfg.expected[key])
        }
      }),
  )
})

describe('searchSelector', () => {
  test('return the search parameter', () => {
    expect(
      searchSelector({
        router: { location: { search: '?search=testing' } },
      } as ReduxState),
    ).toBe('testing')
  })
  test('always return a string', () => {
    expect(
      searchSelector({
        router: { location: { search: '' } },
      } as ReduxState),
    ).toBe('')
  })
})

describe('absURLSelector', () => {
  const sel = absURLSelector({
    router: { location: { pathname: '/base' } },
  } as ReduxState)
  test('clean urls', () => {
    expect(sel('/foo/.')).toBe('http://localhost/foo')
    expect(sel('foo/././/')).toBe('http://localhost/base/foo')
  })
  test('respect absolute urls', () => {
    expect(sel('/foo')).toBe('http://localhost/foo')
  })
  test('join relative urls', () => {
    expect(sel('foo')).toBe('http://localhost/base/foo')
    expect(sel('foo/')).toBe('http://localhost/base/foo')
  })

  test('handle .. appropriately', () => {
    const check = (base: string, path: string, expected: string) =>
      expect(
        absURLSelector({
          router: { location: { pathname: base } },
        } as ReduxState)(path),
      ).toBe(expected)

    check('/foo/bar', '..', 'http://localhost/foo')
    check('/foo/bar/', '..', 'http://localhost/foo')
    check('/foo/bar', '../baz', 'http://localhost/foo/baz')
    check('/foo/bar/', '../baz/', 'http://localhost/foo/baz')
    check('/foo/bar/', '../baz/../bin', 'http://localhost/foo/bin')
  })
})
