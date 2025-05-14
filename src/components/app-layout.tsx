import { AppSidebar } from "@/components/app-sidebar"
import { useTheme } from "@/components/theme-provider"
import { SidebarInset } from "@/components/ui/sidebar"
import { cn } from "@/lib/utils"
import { useSidebar } from "@/components/ui/sidebar"

interface AppLayoutProps {
  children: React.ReactNode
}

export function AppLayout({ children }: AppLayoutProps) {
  const { theme } = useTheme()
  const { state } = useSidebar()
  
  return (
    <div className={cn(
      "flex h-screen w-full overflow-hidden transition-colors duration-300 ease-in-out",
      state === "expanded" ? "bg-sidebar" : "bg-background"
    )}>
      <AppSidebar />
      {state === "expanded" ? (
        <SidebarInset className={cn(
          "bg-background transition-colors duration-300 ease-in-out",
          theme === "dark" && "text-foreground"
        )}>
          <div className="p-4">
            {children}
          </div>
        </SidebarInset>
      ) : (
        <div className={cn(
          "flex-1 transition-colors duration-300 ease-in-out",
          theme === "dark" && "text-foreground"
        )}>
          <div className="p-4">
            {children}
          </div>
        </div>
      )}
    </div>
  )
}