import { createRouter, createWebHistory } from 'vue-router'
import EventsPage from '../pages/EventsPage.vue'
import ProfilePage from '../pages/ProfilePage.vue'
export default createRouter({ history: createWebHistory(), routes: [{ path: '/', redirect: '/events' }, { path: '/events', component: EventsPage }, { path: '/profile', component: ProfilePage }] })
