import { createRouter, createWebHashHistory } from "vue-router"
import DashboardPrincipal from "@/views/dashboard-principal.vue"
import HomeView from "@/views/HomeView.vue"
import PaginaPrincipal from "@/views/pagina-principal.vue"
import SupportAlert from "@/views/support-alert.vue"
import WarningAlert from "@/views/warning-alert.vue"

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: "/", name: "home", component: HomeView },
    { path: "/warning-alert", name: "warning-alert", component: WarningAlert },
    { path: "/support-alert", name: "support-alert", component: SupportAlert },
    { path: "/pagina-principal", name: "pagina-principal", component: PaginaPrincipal },
    { path: "/dashboard-principal", name: "dashboard-principal", component: DashboardPrincipal },
  ],
})

export default router
