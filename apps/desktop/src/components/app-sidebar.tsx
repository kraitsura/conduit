import { Clock, LayoutDashboard, Settings, Lightbulb, BarChart } from "lucide-react"
import { ThemeToggle } from "@/components/theme-toggle"
import { ThemeSetToggle } from "@/components/theme-sets/theme-set-toggle"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarSeparator,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { useSidebar } from "@/components/ui/sidebar"
import { cn } from "@/lib/utils"

export function AppSidebar() {
  const { state } = useSidebar()
  
  return (
    <Sidebar
      variant={state === "expanded" ? "inset" : "sidebar"}
      className={cn(
        "h-screen transition-all duration-300 ease-in-out origin-left",
        state === "expanded" ? "bg-sidebar" : "bg-background"
      )}
      style={{
        "--sidebar-width": "160px",
        "--sidebar-width-icon": "40px"
      } as React.CSSProperties}
      collapsible="icon"
    >
      <SidebarHeader>
        <div className="flex items-center px-2">
            <h2 className={cn(
              "text-lg font-semibold transition-all duration-300 absolute",
              state === "expanded" ? "text-sidebar-foreground opacity-100" : "text-foreground opacity-0",
              "transform"            
            )}>
              Conduit
            </h2>
          <div className="flex items-center ml-auto relative z-10">
            <Tooltip>
              <TooltipTrigger asChild>
                <SidebarTrigger className={cn(
                  "transition-all duration-300 w-10 h-10 flex items-center justify-center",
                  state === "expanded" 
                    ? "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground ml-auto" 
                    : "hover:bg-sidebar-accent -ml-4"
                )} />
              </TooltipTrigger>
              <TooltipContent side="right" hidden={state !== "collapsed"}>
                Toggle Sidebar
              </TooltipContent>
            </Tooltip>
          </div>
        </div>
      </SidebarHeader>
      <SidebarSeparator className={cn(
        "transition-colors duration-300",
        state === "collapsed" ? "bg-border opacity-0" : "opacity-100"
      )} />
      <SidebarContent className="flex flex-col items-center">
        <SidebarMenu className="w-full">
          <SidebarMenuItem className="flex justify-center">
            <SidebarMenuButton 
              isActive={true} 
              tooltip="Dashboard"
              className={cn(
                "transition-all duration-300 relative h-10",
                "text-foreground hover:text-accent-foreground",
                "data-[active=true]:text-primary",
                state === "expanded" 
                  ? "w-full" 
                  : "justify-center w-8"
              )}
            >
              <LayoutDashboard className={cn(
                "w-5 h-5",
                state === "collapsed" && "ml-2"
              )} />
              <span className={cn(
                "transition-opacity duration-300",
                state === "collapsed" && "opacity-0"
              )}>Dashboard</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem className="flex justify-center">
            <SidebarMenuButton 
              tooltip="Focus Timer"
              className={cn(
                "transition-all duration-300 relative h-10",
                "text-foreground hover:text-accent-foreground",
                state === "expanded" 
                  ? "w-full" 
                  : "justify-center w-8"
              )}
            >
              <Clock className={cn(
                "w-5 h-5",
                state === "collapsed" && "ml-2"
              )} />
              <span className={cn(
                "transition-opacity duration-300",
                state === "collapsed" && "opacity-0"
              )}>Focus Timer</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem className="flex justify-center">
            <SidebarMenuButton 
              tooltip="Tasks"
              className={cn(
                "transition-all duration-300 relative h-10",
                "text-foreground hover:text-accent-foreground",
                state === "expanded" 
                  ? "w-full" 
                  : "justify-center w-8"
              )}
            >
              <Lightbulb className={cn(
                "w-5 h-5",
                state === "collapsed" && "ml-2"
              )} />
              <span className={cn(
                "transition-opacity duration-300",
                state === "collapsed" && "opacity-0"
              )}>Tasks</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem className="flex justify-center">
            <SidebarMenuButton 
              tooltip="Analytics"
              className={cn(
                "transition-all duration-300 relative h-10",
                "text-foreground hover:text-accent-foreground",
                state === "expanded" 
                  ? "w-full" 
                  : "justify-center w-8"
              )}
            >
              <BarChart className={cn(
                "w-5 h-5",
                state === "collapsed" && "ml-2"
              )} />
              <span className={cn(
                "transition-opacity duration-300",
                state === "collapsed" && "opacity-0"
              )}>Analytics</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
          <SidebarMenuItem className="flex justify-center">
            <SidebarMenuButton 
              tooltip="Settings"
              className={cn(
                "transition-all duration-300 relative h-10",
                "text-foreground hover:text-accent-foreground",
                state === "expanded" 
                  ? "w-full" 
                  : "justify-center w-8"
              )}
            >
              <Settings className={cn(
                "w-5 h-5",
                state === "collapsed" && "ml-2"
              )} />
              <span className={cn(
                "transition-opacity duration-300",
                state === "collapsed" && "opacity-0"
              )}>Settings</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarContent>
      <SidebarFooter className={cn(
        "mt-auto flex flex-col items-center transition-opacity duration-300 py-4 gap-2",
        state === "collapsed" && "opacity-50"
      )}>
        {state === "collapsed" ? (
          <>
            <ThemeSetToggle compact />
            <ThemeToggle />
          </>
        ) : (
          <div className="flex gap-2">
            <ThemeSetToggle />
            <ThemeToggle />
          </div>
        )}
      </SidebarFooter>
    </Sidebar>
  )
} 