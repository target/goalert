import React from 'react'

import { Text, TextProps } from './Themed'

export function MonoText(props: TextProps): JSX.Element {
  return <Text {...props} style={[props.style, { fontFamily: 'space-mono' }]} />
}
