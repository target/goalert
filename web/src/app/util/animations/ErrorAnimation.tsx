import React from 'react'
import Lottie from 'react-lottie'
import * as animationData from './error-sad-face.json'

interface ErrorAnimationProps {
  autoplay?: boolean
  isStopped?: boolean
  loop?: boolean
}

export default function ErrorAnimation(
  props: ErrorAnimationProps,
): JSX.Element {
  const defaultOptions = {
    loop: props.loop,
    autoplay: props.autoplay,
    animationData: animationData,
    rendererSettings: {
      preserveAspectRatio: 'xMidYMid slice',
    },
  }

  return (
    <Lottie
      options={defaultOptions}
      isStopped={props.isStopped}
      height={175}
      width={175}
      style={{
        margin: '-40px auto auto auto',
      }}
    />
  )
}
