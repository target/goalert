import { gql } from 'urql'
import { DocumentNode, print } from 'graphql'

import {
  queryByName,
  fieldAlias,
  mapInputVars,
  mergeFields,
  prefixQuery,
} from './graphql'

const expectEqual = (a: DocumentNode, b: DocumentNode) => expect(print(a)).toBe(print(b))

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
  const check = (name: string, arg: string, query: DocumentNode, expected: DocumentNode) =>
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

describe('prefixQuery', () => {
  const check = (name: string, arg: string, query: DocumentNode, expected: DocumentNode) =>
    test(name, () =>
      expect(print(prefixQuery(query, 'q0_'))).toBe(print(expected)),
    )

  check(
    'should prefix query and variables',
    '',
    gql`
      query ($id: ID!, $id2: ID!) {
        user(id: $id) {
          id
          name
        }
        user2: user(id: $id2) {
          id
          name
        }
      }
    `,
    gql`
      query ($q0_id: ID!, $q0_id2: ID!) {
        q0_user: user(id: $q0_id) {
          id
          name
        }
        q0_user2: user(id: $q0_id2) {
          id
          name
        }
      }
    `,
  )
})

describe('mapInputVars', () => {
  const check = (name: string, arg: Record<string, string>, query: DocumentNode, expected: DocumentNode) =>
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
  const check = (name: string, query1: DocumentNode, query2: DocumentNode, expected: DocumentNode) =>
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
