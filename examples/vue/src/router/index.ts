import { createRouter, createWebHistory } from "vue-router";
import LoginView from "../views/LoginView.vue";
import TodoView from "../views/TodoView.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "login",
      component: LoginView
    },
    {
      path: "/todo",
      name: "todo",
      component: TodoView
    }
  ]
});

export default router;
