import Vue from 'vue';
import Router from 'vue-router';
import Login from '../pages/Login';
import Main from '../pages/Main';

Vue.use(Router);

export default new Router({
  routes: [
    {
      path: '/',
      name: 'Main',
      component: Main,
    },
    {
      path: '/login',
      name: 'Login',
      component: Login,
    },
  ],
});
