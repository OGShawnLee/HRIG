import { createRouter, createWebHashHistory } from "vue-router"
import HomeView from "@/views/HomeView.vue"
import SearchView from "@/views/SearchView.vue"
import DashboardPrincipal from "@/views/dashboard-principal.vue"
import DashboardSecundario from "@/views/dashboard-secundario.vue"
import SupportAlert from "@/views/support-alert.vue"
import WarningAlert from "@/views/warning-alert.vue"

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/", name: "home", component: HomeView },
    { path: "/search", name: "search", component: SearchView },
    { path: "/dashboard-principal", name: "dashboard-principal", component: DashboardPrincipal },
    { path: "/other-stats", name: "other-stats", component: DashboardSecundario },
    { path: "/secondary-stats", redirect: { name: "other-stats" } },
    { path: "/support-alert", name: "support-alert", component: SupportAlert },
    { path: "/warning-alert", name: "warning-alert", component: WarningAlert },
    { path: "/pagina-principal", redirect: { name: "home" } },
  ],
})

export default router
