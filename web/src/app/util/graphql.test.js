import gql from 'graphql-tag'
import { print } from 'graphql/language/printer'

import { queryByName, fieldAlias, mapInputVars, mergeFields } from './graphql'

const expectEqual = (a, b) => expect(print(a)).toBe(print(b))

describe('queryByName', () => {
  it('should fetch a single query', () => {
    const query = gql`
      query A {
        test1
      }
      query B {
        test2
      }
    `

    const expected = gql`
      query A {
        test1
      }
    `
    expectEqual(queryByName(query, 'A'), expected)
  })
})

describe('fieldAlias', () => {
  const check = (name, arg, query, expected) =>
    test(name, () =>
      expect(print(fieldAlias(query, arg))).toBe(print(expected)),
    )

  check(
    'should rename existing alias',
    'bar',
    gql`
      query Test {
        foo: get {
          id
        }
      }
    `,
    gql`
      query Test {
        bar: get {
          id
        }
      }
    `,
  )

  check(
    'should add new alias',
    'bar',
    gql`
      query Test {
        get {
          id
        }
      }
    `,
    gql`
      query Test {
        bar: get {
          id
        }
      }
    `,
  )
})

describe('mapInputVars', () => {
  const check = (name, arg, query, expected) =>
    test(name, () =>
      expect(print(mapInputVars(query, arg))).toBe(print(expected)),
    )

  check(
    'should handle no changes',
    {},
    gql`
      query Test($id: ID!) {
        thing(id: $id) {
          id
          name
        }
      }
    `,
    gql`
      query Test($id: ID!) {
        thing(id: $id) {
          id
          name
        }
      }
    `,
  )

  check(
    'should rename variables',
    { id: 'bob' },
    gql`
      query Test($id: ID!) {
        thing(id: $id) {
          id
          name
        }
      }
    `,
    gql`
      query Test($bob: ID!) {
        thing(id: $bob) {
          id
          name
        }
      }
    `,
  )

  check(
    'should handle extra mappings',
    { id: 's1', foo: 'bar' },
    gql`
      query Test($id: ID!) {
        thing(id: $id) {
          id
          name
        }
      }
    `,
    gql`
      query Test($s1: ID!) {
        thing(id: $s1) {
          id
          name
        }
      }
    `,
  )
})

describe('mergeFields', () => {
  const check = (name, query1, query2, expected) =>
    test(name, () =>
      expect(print(mergeFields(query1, query2))).toBe(print(expected)),
    )

  check(
    'should merge fields',
    gql`
      query First {
        foo
        bar {
          baz
        }
        bin: ok
      }
    `,
    gql`
      query Second {
        foo2
        bar2 {
          baz2
        }
        bin2: ok2
      }
    `,
    gql`
      query First {
        foo
        bar {
          baz
        }
        bin: ok
        foo2
        bar2 {
          baz2
        }
        bin2: ok2
      }
    `,
  )

  check(
    'should merge input variables',
    gql`
      query First($id: ID!) {
        foo
        bar(id: $id) {
          baz
        }
        bin: ok
      }
    `,
    gql`
      query Second($input: ThatInput) {
        foo2
        bar2(input: $input) {
          baz2
        }
        bin2: ok2
      }
    `,
    gql`
      query First($id: ID!, $input: ThatInput) {
        foo
        bar(id: $id) {
          baz
        }
        bin: ok
        foo2
        bar2(input: $input) {
          baz2
        }
        bin2: ok2
      }
    `,
  )
})
