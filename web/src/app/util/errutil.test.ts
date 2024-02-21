import { GraphQLError } from 'graphql'
import { splitErrorsByPath } from './errutil'

describe('splitErrorsByPath', () => {
  it('should split errors by path', () => {
    const resp = [
      {
        message: 'test1',
        path: ['foo', 'bar', 'dest', 'type'],
        extensions: {
          code: 'INVALID_INPUT_VALUE',
        },
      },
      {
        message: 'test2',
        path: ['foo', 'bar', 'dest', 'values', 'example-field'],
        extensions: {
          code: 'INVALID_INPUT_VALUE',
        },
      },
    ] as unknown as GraphQLError[]

    const [inputFieldErrors, otherErrors] = splitErrorsByPath(resp, [
      'foo.bar.dest.type',
      'foo.bar.dest.values.example-field',
    ])

    expect(inputFieldErrors).toHaveLength(2)
    expect(inputFieldErrors[0].message).toEqual('test1')
    expect(inputFieldErrors[0].path.join('.')).toEqual('foo.bar.dest.type')
    expect(inputFieldErrors[1].message).toEqual('test2')
    expect(inputFieldErrors[1].path.join('.')).toEqual(
      'foo.bar.dest.values.example-field',
    )
    expect(otherErrors).toHaveLength(0)
  })
})
