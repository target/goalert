import withPathPrefix from './withPathPrefix'

test('it should prepend the prefix to absolute paths', () => {
  expect(withPathPrefix('/rotations/123', '/goalert')).toBe(
    '/goalert/rotations/123',
  )
  expect(withPathPrefix('/rotations/123', '/goalert/')).toBe(
    '/goalert/rotations/123',
  )
  expect(withPathPrefix('/', '/goalert')).toBe('/goalert/')
})

test('it should not change paths when there is no prefix', () => {
  expect(withPathPrefix('/rotations/123', '')).toBe('/rotations/123')
})

test('it should not change external or relative URLs', () => {
  expect(withPathPrefix('https://example.com/foo', '/goalert')).toBe(
    'https://example.com/foo',
  )
  expect(withPathPrefix('//example.com/foo', '/goalert')).toBe(
    '//example.com/foo',
  )
  expect(withPathPrefix('mailto:a@example.com', '/goalert')).toBe(
    'mailto:a@example.com',
  )
  expect(withPathPrefix('foo/bar', '/goalert')).toBe('foo/bar')
})
