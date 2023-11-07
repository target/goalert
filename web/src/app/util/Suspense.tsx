// Implementation sourced from:
// https://github.com/asifvora/react-current-page-fallback
import React, {
  ReactNode,
  Suspense,
  Fragment,
  FC,
  createContext,
  useCallback,
  useContext,
  useMemo,
  useEffect,
  useState,
} from 'react'
import { LinearProgress } from '@mui/material'

export interface FallbackContextType {
  updateFallback: (fallbackElement: ReactNode) => void
}

export const FallbackContext = createContext<FallbackContextType>({
  updateFallback: () => {},
})

interface FallbackProviderProps {
  children: ReactNode
}

export const FallbackProvider: FC<FallbackProviderProps> = ({ children }) => {
  const [fallback, setFallback] = useState<ReactNode>(null)

  const updateFallback = useCallback((fallbackElement: ReactNode) => {
    setFallback(() => fallbackElement)
  }, [])

  const renderChildren = useMemo(() => children, [children])

  return (
    <FallbackContext.Provider value={{ updateFallback }}>
      <Suspense
        fallback={
          <Fragment>
            <LinearProgress />
            {fallback}
          </Fragment>
        }
      >
        {renderChildren}
      </Suspense>
    </FallbackContext.Provider>
  )
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const usePageRoute = (): any => {
  const { updateFallback } = useContext(FallbackContext)

  const onLoad = useCallback(
    (component: ReactNode) => {
      if (component === undefined) {
        component = null
      }
      updateFallback(component)
    },
    [updateFallback],
  )

  return { onLoad }
}

interface PageWrapperProps {
  children?: ReactNode
}

export const FallbackPageWrapper: FC<PageWrapperProps> = ({
  children,
}: PageWrapperProps) => {
  const { onLoad } = usePageRoute()

  const render = useMemo(() => children, [children])

  useEffect(() => {
    onLoad(render)
  }, [onLoad, render])

  return render
}
