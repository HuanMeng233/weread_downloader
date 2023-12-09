import { createRouter, createWebHistory } from "vue-router";
import Index from "../views/Index.vue";
import Home from "../views/Home.vue";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      name: "index",
      component: Index,
    },
    {
      path: "/home",
      name: "home",
      component: Home,
    },
  ],
});

export default router;
