import { urlParamSelector } from '../selectors'
import { setURLParam, resetURLParams } from './main'
import { useSelector, useDispatch } from 'react-redux'

export function useURLParam(name, _default) {
  const dispatch = useDispatch()
  const urlParam = useSelector(urlParamSelector)
  const value = urlParam(name, _default)
  const setValue = value => dispatch(setURLParam(name, value, _default))

  return [value, setValue]
}

export function useResetURLParams(...keys) {
  const dispatch = useDispatch()
  return () => dispatch(resetURLParams(...keys))
}
