import React from 'react'
import { it, expect } from 'jest-expo'
import renderer from 'react-test-renderer'

import { MonoText } from '../StyledText'

it(`renders correctly`, () => {
  const tree = renderer.create(<MonoText>Snapshot test!</MonoText>).toJSON()

  expect(tree).toMatchSnapshot()
})
