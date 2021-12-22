import { getURLParam, sanitizeURLParam, Value } from './hooks'

interface GetParamTest {
  desc: string
  name: string
  defaultValue: Value
  expected: Value
}

describe('getURLParam', () => {
  const init = 'a=str&b=3&c=1&d=0&e=e&e=e&e=ee'
  const q = new URLSearchParams(init)

  function check(x: GetParamTest): void {
    it(x.desc, () => {
      expect(getURLParam(q, x.name, x.defaultValue)).toEqual(x.expected)
    })
  }

  check({
    desc: 'string value',
    name: 'a',
    defaultValue: 'aaa',
    expected: 'str',
  })

  check({
    desc: 'mising string value',
    name: 'zzz',
    defaultValue: 'aaa',
    expected: 'aaa',
  })

  check({
    desc: 'string multi-value',
    name: 'e',
    defaultValue: ['extra'],
    expected: ['e', 'e', 'ee'],
  })

  check({
    desc: 'missing multi-value',
    name: 'zzz',
    defaultValue: ['extra'],
    expected: ['extra'],
  })

  check({
    desc: 'number value 1',
    name: 'b',
    defaultValue: 4,
    expected: 3,
  })

  check({
    desc: 'number value 2',
    name: 'c',
    defaultValue: 4,
    expected: 1,
  })

  check({
    desc: 'number value 3',
    name: 'd',
    defaultValue: 4,
    expected: 0,
  })

  check({
    desc: 'missing number value',
    name: 'zzz',
    defaultValue: 4,
    expected: 4,
  })

  check({
    desc: 'bool value (true)',
    name: 'c',
    defaultValue: false,
    expected: true,
  })

  check({
    desc: 'bool value (false)',
    name: 'd',
    defaultValue: false,
    expected: false,
  })

  check({
    desc: 'missing bool value',
    name: 'zzz',
    defaultValue: true,
    expected: true,
  })
})

interface SanitizeTest {
  desc: string
  val: Value
  expected: string | string[]
}

describe('sanitizeURLParam', () => {
  function check(x: SanitizeTest): void {
    it(x.desc, () => {
      expect(sanitizeURLParam(x.val)).toEqual(x.expected)
    })
  }

  check({
    desc: 'string value',
    val: 'str',
    expected: 'str',
  })

  check({
    desc: 'empty string',
    val: '',
    expected: '',
  })

  check({
    desc: 'trim string',
    val: '  ',
    expected: '',
  })

  check({
    desc: 'multi string',
    val: [' hello', '', ' world ', '   '],
    expected: ['hello', 'world'],
  })

  check({
    desc: 'bool value true',
    val: true,
    expected: '1',
  })

  check({
    desc: 'bool value false',
    val: false,
    expected: '',
  })

  check({
    desc: 'num value 1',
    val: 1,
    expected: '1',
  })

  check({
    desc: 'num value 2',
    val: 0,
    expected: '0',
  })
})
