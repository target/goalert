import AdminRouter from './AdminRouter'

export default [
  {
    nav: false,
    title: 'Config',
    path: '/admin',
    component: AdminRouter,
  },
  {
    nav: false,
    title: 'User Management',
    path: '/admin/user-management',
    component: AdminRouter,
  },
  {
    nav: false,
    title: 'System',
    path: '/admin/system',
    component: AdminRouter,
  },
  {
    nav: false,
    title: 'Reports',
    path: '/admin/reports',
    component: AdminRouter,
  },
]
