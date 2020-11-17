import React from 'react'
import { Ionicons } from '@expo/vector-icons'
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs'
import { createStackNavigator } from '@react-navigation/stack'

import Colors from '../constants/Colors'
import useColorScheme from '../hooks/useColorScheme'
import HomeScreen from '../screens/HomeScreen'
import SettingsScreen from '../screens/SettingsScreen'
import { BottomTabParamList, TabOneParamList, TabTwoParamList } from '../types'

// You can explore the built-in icon families and icons on the web at:
// https://icons.expo.fyi/
function TabBarIcon(props: { name: string; color: string }) {
  return <Ionicons size={30} style={{ marginBottom: -3 }} {...props} />
}

// Each tab has its own navigation stack, you can read more about this pattern here:
// https://reactnavigation.org/docs/tab-based-navigation#a-stack-navigator-for-each-tab
const TabOneStack = createStackNavigator<TabOneParamList>()

function HomeScreenNavigator() {
  return (
    <TabOneStack.Navigator>
      <TabOneStack.Screen
        name='Home'
        component={HomeScreen}
        options={{ headerTitle: 'Home' }}
      />
    </TabOneStack.Navigator>
  )
}

const TabTwoStack = createStackNavigator<TabTwoParamList>()

function SettingsScreenNavigator() {
  return (
    <TabTwoStack.Navigator>
      <TabTwoStack.Screen
        name='Settings'
        component={SettingsScreen}
        options={{ headerTitle: 'Settings' }}
      />
    </TabTwoStack.Navigator>
  )
}

const BottomTab = createBottomTabNavigator<BottomTabParamList>()

export default function BottomTabNavigator() {
  const colorScheme = useColorScheme()

  return (
    <BottomTab.Navigator
      initialRouteName='Home'
      tabBarOptions={{ activeTintColor: Colors[colorScheme].tint }}
    >
      <BottomTab.Screen
        name='Home'
        component={HomeScreenNavigator}
        options={{
          tabBarIcon: ({ color }) => (
            <TabBarIcon name='ios-code' color={color} />
          ),
        }}
      />
      <BottomTab.Screen
        name='Settings'
        component={SettingsScreenNavigator}
        options={{
          tabBarIcon: ({ color }) => (
            <TabBarIcon name='ios-code' color={color} />
          ),
        }}
      />
    </BottomTab.Navigator>
  )
}
