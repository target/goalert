import { createSelector } from 'reselect'
import joinURL from '../util/joinURL'
import { memoize } from 'lodash-es'

export const urlQuerySelector = state => state.router.location.search
export const urlPathSelector = state => state.router.location.pathname

export const urlSearchParamsSelector = createSelector(
  urlQuerySelector,

  query => new URLSearchParams(query),
)

export const urlParamSelector = createSelector(
  urlSearchParamsSelector,
  params =>
    memoize((name, _default = null) => {
      if (!params.has(name)) return _default

      if (Array.isArray(_default)) return params.getAll(name)
      if (typeof _default === 'boolean') return Boolean(params.get(name))
      if (typeof _default === 'number') return +params.get(name)

      return params.get(name)
    }),
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

export const absURLSelector = createSelector(urlPathSelector, base =>
  memoize(
    path =>
      path && (path.startsWith('/') ? joinURL(path) : joinURL(base, path)),
  ),
)
