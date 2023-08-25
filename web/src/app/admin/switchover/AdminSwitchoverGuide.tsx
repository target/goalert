import React from 'react'
import swoGuide from '../../../../../docs/switchover.md'

import Markdown from '../../util/Markdown'

export default function AdminSwitchoverGuide(): JSX.Element {
  return <Markdown value={swoGuide} />
}
