import { Outlet, createRootRoute } from "@tanstack/react-router"
import { useEffect } from "react"

import { AppLayout } from "@/components/app-layout"
import { initializeChatStore } from "@/features/chat/controller"

const RootLayout = () => {
  useEffect(() => {
    initializeChatStore()
  }, [])

  return (
    <AppLayout>
      <Outlet />
    </AppLayout>
  )
}

export const Route = createRootRoute({ component: RootLayout })
