import { GraphQLError } from 'graphql'
import { getInputFieldErrors } from './errutil'

describe('getInputFieldErrors', () => {
  it('should split errors by path', () => {
    const resp = {
      name: 'ignored',
      message: 'ignored',
      graphQLErrors: [
        {
          message: 'test1',
          path: ['foo', 'bar', 'dest', 'type'],
          extensions: {
            code: 'INVALID_DESTINATION_TYPE',
          },
        },
        {
          message: 'test2',
          path: ['foo', 'bar', 'dest', 'values', 'example-field'],
          extensions: {
            code: 'INVALID_DESTINATION_FIELD_VALUE',
          },
        },
      ] as unknown as GraphQLError[],
    }

    const [inputFieldErrors, otherErrors] = getInputFieldErrors(
      ['foo.bar.dest.type', 'foo.bar.dest.values.example-field'],
      resp,
    )

    expect(inputFieldErrors).toHaveLength(2)
    expect(inputFieldErrors[0].message).toEqual('test1')
    expect(inputFieldErrors[0].path).toEqual('foo.bar.dest.type')
    expect(inputFieldErrors[1].message).toEqual('test2')
    expect(inputFieldErrors[1].path).toEqual(
      'foo.bar.dest.values.example-field',
    )
    expect(otherErrors).toHaveLength(0)
  })
})
