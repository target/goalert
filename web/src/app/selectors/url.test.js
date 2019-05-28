import { urlParamSelector, searchSelector, absURLSelector } from './url'

describe('urlParamSelector', () => {
  ;[
    { search: '?search=foo', expected: { search: 'foo' } },
    { search: '', expected: {} },
    { search: '?&search=foo', expected: { search: 'foo' } },
    { search: '?search=foo&&&', expected: { search: 'foo' } },
    { search: '?foo=bar&bin=baz', expected: { foo: 'bar', bin: 'baz' } },
    { search: '?search=asdf%26%3D', expected: { search: 'asdf&=' } },
  ].forEach(cfg =>
    test(cfg.search || '(empty)', () => {
      const res = urlParamSelector({
        router: {
          location: { search: cfg.search },
        },
      })
      for (let key in cfg.expected) {
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
      }),
    ).toBe('testing')
  })
  test('always return a string', () => {
    expect(
      searchSelector({
        router: { location: { search: '' } },
      }),
    ).toBe('')
  })
})

describe('absURLSelector', () => {
  const sel = absURLSelector({
    router: { location: { pathname: '/base' } },
  })
  test('clean urls', () => {
    expect(sel('/foo/.')).toBe('/foo')
    expect(sel('foo/././/')).toBe('/base/foo')
  })
  test('respect absolute urls', () => {
    expect(sel('/foo')).toBe('/foo')
  })
  test('join relative urls', () => {
    expect(sel('foo')).toBe('/base/foo')
    expect(sel('foo/')).toBe('/base/foo')
  })

  test('handle .. appropriately', () => {
    const check = (base, path, expected) =>
      expect(
        absURLSelector({ router: { location: { pathname: base } } })(path),
      ).toBe(expected)

    check('/foo/bar', '..', '/foo')
    check('/foo/bar/', '..', '/foo')
    check('/foo/bar', '../baz', '/foo/baz')
    check('/foo/bar/', '../baz/', '/foo/baz')
    check('/foo/bar/', '../baz/../bin', '/foo/bin')
  })
})
