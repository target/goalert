import {
  DocumentNode,
  FieldNode,
  DefinitionNode,
  Kind,
  NameNode,
  OperationDefinitionNode,
  SelectionNode,
  VariableDefinitionNode,
} from 'graphql'

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
export function queryByName(doc: DocumentNode, name: string): DocumentNode {
  const def = doc.definitions.find(
    (def) => 'name' in def && def.name?.value === name,
  )
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
export function fieldAlias(doc: DocumentNode, aliasName: string): DocumentNode {
  if (doc.definitions.length > 1) {
    throw new Error(
      `found ${doc.definitions.length} query definitions, but expected 1`,
    )
  }
  const def = doc.definitions[0] as OperationDefinitionNode
  if (def.selectionSet.selections.length > 1) {
    throw new Error(
      `found ${def.selectionSet.selections.length} fields, but expected 1`,
    )
  }
  const sel = def.selectionSet.selections[0] as SelectionNode

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
              alias: {
                kind: 'Name' as Kind.NAME,
                value: aliasName,
              },
            } as FieldNode,
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
export function prefixQuery(doc: DocumentNode, prefix: string): DocumentNode {
  const mapVarName = (name: { value: string }): NameNode => ({
    ...name,
    value: prefix + name.value,
    kind: Kind.NAME,
  })
  const mapSelName = (sel: FieldNode): FieldNode => {
    if ('alias' in sel && sel.alias) {
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
        kind: 'Name' as Kind.NAME,
        value: prefix + sel.name.value,
      },
    }
  }
  return {
    ...doc,
    definitions: doc.definitions.map((def) => {
      if (def.kind !== 'OperationDefinition') return def
      return {
        ...def,
        variableDefinitions: def.variableDefinitions?.map(
          (vDef: VariableDefinitionNode) => ({
            ...vDef,
            variable: {
              ...vDef.variable,
              name: mapVarName(vDef.variable.name),
            },
          }),
        ),
        selectionSet: {
          ...def.selectionSet,
          selections: def.selectionSet.selections.map((sel: SelectionNode) => ({
            ...mapSelName(sel as FieldNode),
            arguments: (sel as FieldNode).arguments?.map((arg) => ({
              ...arg,
              value:
                arg.value.kind !== 'Variable'
                  ? arg.value
                  : { ...arg.value, name: mapVarName(arg.value.name) },
            })),
          })),
        },
      } as DefinitionNode
    }),
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
export function mapInputVars(
  doc: DocumentNode,
  mapVars: Record<string, string> = {},
): DocumentNode {
  const mapName = (name: { value: string }): NameNode => ({
    ...name,
    value: mapVars[name.value] || name.value,
    kind: Kind.NAME,
  })

  return {
    ...doc,
    definitions: doc.definitions.map((def) => {
      if (def.kind !== 'OperationDefinition') return def
      return {
        ...def,
        variableDefinitions: def.variableDefinitions?.map(
          (vDef: VariableDefinitionNode) => ({
            ...vDef,
            variable: {
              ...vDef.variable,
              name: mapName(vDef.variable.name),
            },
          }),
        ),
        selectionSet: {
          ...def.selectionSet,
          selections: def.selectionSet.selections.map((sel: SelectionNode) => ({
            ...sel,
            arguments: (sel as FieldNode).arguments?.map((arg) => ({
              ...arg,
              value:
                arg.value.kind !== 'Variable'
                  ? arg.value
                  : { ...arg.value, name: mapName(arg.value.name) },
            })),
          })),
        },
      } as DefinitionNode
    }),
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
export function mergeFields(
  doc: DocumentNode,
  newQuery: DocumentNode,
): DocumentNode {
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

  const def = doc.definitions[0] as OperationDefinitionNode
  const newDef = newQuery.definitions[0] as OperationDefinitionNode
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
