import './index.css'
import DeepWorkAssistant from './components/Deepwork'
import { ThemeProvider } from '@/components/theme-provider'
import { ThemeSetProvider } from '@/components/theme-sets/theme-set-provider'
import { AppLayout } from '@/components/app-layout'
import { SidebarProvider } from '@/components/ui/sidebar'

function App() {
  return (
    <ThemeProvider defaultTheme="dark">
      <ThemeSetProvider defaultThemeSet="default" defaultAccent="indigo">
        <SidebarProvider defaultOpen={true}>
          <AppLayout>
            <DeepWorkAssistant />
          </AppLayout>
        </SidebarProvider>
      </ThemeSetProvider>
    </ThemeProvider>
  )
}

export default App