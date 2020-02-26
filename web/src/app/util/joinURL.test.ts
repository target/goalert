import joinURL from './joinURL'

test('it should join URLs correctly', () => {
  expect(joinURL('foo', 'bar')).toBe('foo/bar')
  expect(joinURL('/foo', 'bar')).toBe('/foo/bar')
  expect(joinURL('foo/', 'bar')).toBe('foo/bar')
  expect(joinURL('foo/', '/bar')).toBe('foo/bar')
  expect(joinURL('/foo/', '/bar')).toBe('/foo/bar')
  expect(joinURL('foo', 'bar/')).toBe('foo/bar')
  expect(joinURL('foo', 'bar/', 'baz', 'bin')).toBe('foo/bar/baz/bin')
  expect(joinURL('foo')).toBe('foo')
  expect(joinURL('/foo')).toBe('/foo')
  expect(joinURL()).toBe('')

  expect(joinURL('/foo/bar', '..')).toBe('/foo')

  expect(joinURL('/foo/bar', '..')).toBe('/foo')
  expect(joinURL('/foo/bar/', '..')).toBe('/foo')
  expect(joinURL('/foo/bar', '../baz')).toBe('/foo/baz')
  expect(joinURL('/foo/bar/', '../baz/')).toBe('/foo/baz')
  expect(joinURL('/foo/bar/', '../baz/../bin')).toBe('/foo/bin')

  expect(joinURL('/base', '/foo/.')).toBe('/base/foo')
})
