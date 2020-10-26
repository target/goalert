/**
 * queryByName returns a new GraphQL document with a single query from the provided doc.
 *
 *
 * The following example takes a GraphQL document with 2 defined queries (Foo and Bar)
 * and returns a document with only Bar defined.
 *
 * ```
 * const query = gql`
 *  query Foo {
 *    hello
 *  }
 *  query Bar {
 *    world
 *  }
 * `
 *
 * queryByName(query, 'Foo') === gql`
 *  query Bar {
 *    world
 *  }
 * `
 * ```
 *
 */
export function queryByName(doc, name) {
  const def = doc.definitions.find((def) => def.name.value === name)
  if (!def) throw new Error('no definition found for ' + name)
  return {
    ...doc,
    definitions: [def],
  }
}

/**
 * fieldAlias accepts a GraphQL document with a single-field query and returns
 * a new document with the field aliased to the provided aliasName.
 *
 *
 * The following takes a query for a user, and returns one where the user object
 * will be mapped (aliased) to `data`.
 *
 * const query = gql`
 *  query {
 *    user {
 *      id
 *      name
 *    }
 *  }
 * `
 *
 * fieldAlias(query, 'data') === gql`
 *  query {
 *    data: user {
 *      id
 *      name
 *    }
 *  }
 * `
 */
export function fieldAlias(doc, aliasName) {
  if (doc.definitions.length > 1) {
    throw new Error(
      `found ${doc.definitions.length} query definitions, but expected 1`,
    )
  }
  const def = doc.definitions[0]
  if (def.selectionSet.selections.length > 1) {
    throw new Error(
      `found ${def.selectionSet.selections.length} fields, but expected 1`,
    )
  }
  const sel = def.selectionSet.selections[0]

  return {
    ...doc,
    definitions: [
      {
        ...def,
        selectionSet: {
          ...def.selectionSet,
          selections: [
            {
              ...sel,
              alias: { kind: 'Name', value: aliasName },
            },
          ],
        },
      },
    ],
  }
}

/**
 * prefixQuery accepts a GraphQL document and a prefix.
 * All input variables and selection fields will be aliased with the provided prefix.
 *
 *
 * The following example takes a query for 2 users and prefixes with `q0_`
 *
 * const query = gql`
 *  query($id: ID!, $id2: ID!) {
 *    user(id: $id) { id, name }
 *    user2(id: $id2) { id, name }
 *  }
 * `
 *
 * prefixQuery(query, 'q0_') === gql`
 *  query($q0_id: ID!, $q0_id2: ID!) {
 *    q0_user: user(id: $q0_id) { id, name }
 *    q0_user2: user(id: $q0_id2) { id, name }
 *  }
 * `
 */
export function prefixQuery(doc, prefix) {
  const mapVarName = (name) => ({
    ...name,
    value: prefix + name.value,
  })
  const mapSelName = (sel) => {
    if (sel.alias) {
      return {
        ...sel,
        alias: {
          ...sel.alias,
          value: prefix + sel.alias.value,
        },
      }
    }

    return {
      ...sel,
      alias: {
        kind: 'Name',
        value: prefix + sel.name.value,
      },
    }
  }
  return {
    ...doc,
    definitions: doc.definitions.map((def) => ({
      ...def,
      variableDefinitions: def.variableDefinitions.map((vDef) => ({
        ...vDef,
        variable: {
          ...vDef.variable,
          name: mapVarName(vDef.variable.name),
        },
      })),
      selectionSet: {
        ...def.selectionSet,
        selections: def.selectionSet.selections.map((sel) => ({
          ...mapSelName(sel),
          arguments: sel.arguments.map((arg) => ({
            ...arg,
            value:
              arg.value.kind !== 'Variable'
                ? arg.value
                : {
                    ...arg.value,
                    name: mapVarName(arg.value.name),
                  },
          })),
        })),
      },
    })),
  }
}

/**
 * mapInputVars accepts a GraphQL document and a map of input variable names.
 * Any input variable matching a key in `mapVars` will be replaced by the value in the map.
 *
 *
 * The following example takes a query for a user, by id, and renames the `id` input param
 * to `userID`.
 *
 * const query = gql`
 *  query($id: ID!) {
 *    user(id: $id) { id, name }
 *  }
 * `
 *
 * mapInputVars(query, {id: 'userID'}) === gql`
 *  query($userID: ID!) {
 *    user(id: $userID) { id, name }
 *  }
 * `
 */
export function mapInputVars(doc, mapVars = {}) {
  const mapName = (name) => ({
    ...name,
    value: mapVars[name.value] || name.value,
  })
  return {
    ...doc,
    definitions: doc.definitions.map((def) => ({
      ...def,
      variableDefinitions: def.variableDefinitions.map((vDef) => ({
        ...vDef,
        variable: {
          ...vDef.variable,
          name: mapName(vDef.variable.name),
        },
      })),
      selectionSet: {
        ...def.selectionSet,
        selections: def.selectionSet.selections.map((sel) => ({
          ...sel,
          arguments: sel.arguments.map((arg) => ({
            ...arg,
            value:
              arg.value.kind !== 'Variable'
                ? arg.value
                : {
                    ...arg.value,
                    name: mapName(arg.value.name),
                  },
          })),
        })),
      },
    })),
  }
}

/**
 * mergeFields takes a GraphQL document, and a second document and merges
 * them into a single query.
 *
 *
 * The following example takes a query for a user and a service, and merges them into one.
 *
 * const query1 = gql`
 *  query($userID: ID!) {
 *    user(id: $userID) { id, name }
 *  }
 * `
 * const query2 = gql`
 *  query($serviceID: ID!) {
 *    service(id: $serviceID) { id, name }
 *  }
 * `
 *
 * mergeFields(query1, query2) === gql`
 *  query($userID: ID!, $serviceID: ID!) {
 *    user(id: $userID) { id, name }
 *    service(id: $serviceID) { id, name }
 *  }
 * `
 */
export function mergeFields(doc, newQuery) {
  if (!doc) return newQuery
  if (doc.definitions.length > 1) {
    throw new Error(
      `found ${doc.definitions.length} query definitions, but expected 1`,
    )
  }
  if (newQuery.definitions.length > 1) {
    throw new Error(
      `found ${newQuery.definitions.length} query definitions in newQuery, but expected 1`,
    )
  }

  const def = doc.definitions[0]
  const newDef = newQuery.definitions[0]
  return {
    ...doc,
    definitions: [
      {
        ...def,
        variableDefinitions: (def.variableDefinitions || []).concat(
          newDef.variableDefinitions || [],
        ),
        selectionSet: {
          ...def.selectionSet,
          selections: def.selectionSet.selections.concat(
            newDef.selectionSet.selections,
          ),
        },
      },
    ],
  }
}
