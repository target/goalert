import React, {
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from 'react'
import { CircularProgress } from '@mui/material'
import _ from 'lodash'

import { DEFAULT_SPIN_DELAY_MS, DEFAULT_SPIN_WAIT_MS } from '../../config'

let uniqID = 0

type SpinContainerProps = {
  loading: boolean
  children: React.ReactNode
}

export default function SpinContainer(
  props: SpinContainerProps,
): React.ReactNode {
  const id = useMemo(() => 'spin_' + uniqID++, [])
  const ref = useRef<HTMLDivElement>(null)
  const [rect, setRect] = useState({ top: 0, left: 0, width: 0, height: 0 })
  const [spin, setSpin] = useState(false)

  useEffect(() => {
    if (props.loading) {
      const t = setTimeout(() => setSpin(true), DEFAULT_SPIN_DELAY_MS)
      return () => clearTimeout(t)
    }

    const t = setTimeout(() => setSpin(false), DEFAULT_SPIN_WAIT_MS)
    return () => clearTimeout(t)
  }, [props.loading])

  let childContainerProps = {}
  if (spin) {
    childContainerProps = {
      'aria-describedby': id,
      'aria-busy': true,
    }
  }

  const handleLayout = (): void => {
    if (!ref.current) return
    const newRect = {
      top: ref.current.offsetTop,
      left: ref.current.offsetLeft,
      width: ref.current.offsetWidth,
      height: ref.current.offsetHeight,
    }
    if (_.isEqual(rect, newRect)) return
    setRect(newRect)
  }
  useEffect(() => {
    window.addEventListener('resize', handleLayout)

    return () => {
      window.removeEventListener('resize', handleLayout)
    }
  }, [])

  useLayoutEffect(handleLayout, [ref.current])

  return (
    <div data-cy-spin-loading={props.loading || spin}>
      {spin && (
        <div
          style={{
            position: 'absolute',
            zIndex: 9999,
            width: rect.width,
            height: rect.height,
            top: rect.top,
            background: 'black',
            opacity: 0.65,
            left: rect.left,
          }}
        >
          <CircularProgress
            style={{
              position: 'absolute',
              left: rect.width / 2 - 20,
              top: rect.height / 2 - 20,
            }}
            id={id}
          />
        </div>
      )}
      <div ref={ref} {...childContainerProps}>
        {props.children}
      </div>
    </div>
  )
}
