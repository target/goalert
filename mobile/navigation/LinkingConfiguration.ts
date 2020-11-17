import * as Linking from 'expo-linking'

export default {
  prefixes: [Linking.makeUrl('/')],
  config: {
    screens: {
      Root: {
        screens: {
          Home: {
            screens: {
              HomeScreen: 'one',
            },
          },
          Settings: {
            screens: {
              SettingsScreen: 'two',
            },
          },
        },
      },
      NotFound: '*',
    },
  },
}
