import { createSelector } from 'reselect'
import joinURL from '../util/joinURL'
import { memoize } from 'lodash-es'
import { History } from 'history'
import { pathPrefix } from '../env'

// TODO: move to ../reducers and define rest of state
export interface ReduxState {
  router: History
}

export const urlQuerySelector = (state: ReduxState) =>
  state.router.location.search
export const urlPathSelector = (state: ReduxState) =>
  state.router.location.pathname
export const urlKeySelector = (state: ReduxState) => state.router.location.key

export const urlSearchParamsSelector = createSelector(
  urlQuerySelector,

  query => new URLSearchParams(query),
)

export const urlParamSelector = createSelector(
  urlSearchParamsSelector,
  params => (
    name: string,
    _default: string | boolean | number | string[] | null = null,
  ) => {
    if (!params.has(name)) return _default

    if (Array.isArray(_default)) return params.getAll(name)
    if (typeof _default === 'boolean') return Boolean(params.get(name))
    if (typeof _default === 'number') return +(params.get(name) as string) // already checked .has()

    return params.get(name)
  },
)

export const searchSelector = createSelector(urlParamSelector, params =>
  params('search', ''),
)

export const alertFilterSelector = createSelector(urlParamSelector, params =>
  params('filter', 'active'),
)

export const alertAllServicesSelector = createSelector(
  urlParamSelector,
  params => params('allServices', false),
)

// absURLSelector will return an absolute URL (including protocol) for the given
// relative or from-root path. It will automatically add any path prefix.
export const absURLSelector = createSelector(urlPathSelector, base =>
  memoize(
    (path: string) =>
      path &&
      location.origin +
        (path.startsWith('/')
          ? joinURL(pathPrefix, path)
          : joinURL(base, path)),
  ),
)
