import { createContext, useContext, useEffect, useState } from "react"

export type ThemeSet = "default" | "patagonia" | "redwood"
export type AccentColor = "indigo" | "rose" | "forest" | "amber" | "teal"

type ThemeSetProviderProps = {
  children: React.ReactNode
  defaultThemeSet?: ThemeSet
  defaultAccent?: AccentColor
}

type ThemeSetContextType = {
  themeSet: ThemeSet
  setThemeSet: (theme: ThemeSet) => void
  accent: AccentColor
  setAccent: (accent: AccentColor) => void
}

const ThemeSetContext = createContext<ThemeSetContextType | undefined>(undefined)

export function ThemeSetProvider({
  children,
  defaultThemeSet = "default",
  defaultAccent = "indigo",
}: ThemeSetProviderProps) {
  const [themeSet, setThemeSet] = useState<ThemeSet>(
    () => (localStorage?.getItem("themeSet") as ThemeSet) || defaultThemeSet
  )
  
  const [accent, setAccent] = useState<AccentColor>(
    () => (localStorage?.getItem("accent") as AccentColor) || defaultAccent
  )

  useEffect(() => {
    const root = window.document.documentElement
    
    // Remove any existing theme-set classes
    root.classList.remove("theme-default", "theme-patagonia", "theme-redwood")
    
    // Add the current theme-set class
    root.classList.add(`theme-${themeSet}`)
    
    // Save to localStorage
    localStorage.setItem("themeSet", themeSet)
  }, [themeSet])
  
  useEffect(() => {
    const root = window.document.documentElement
    
    // Remove any existing accent classes
    root.classList.remove("accent-indigo", "accent-rose", "accent-forest", "accent-amber", "accent-teal")
    
    // Add the current accent class
    root.classList.add(`accent-${accent}`)
    
    // Save to localStorage
    localStorage.setItem("accent", accent)
  }, [accent])

  const value = {
    themeSet,
    setThemeSet,
    accent,
    setAccent,
  }

  return (
    <ThemeSetContext.Provider value={value}>
      {children}
    </ThemeSetContext.Provider>
  )
}

export const useThemeSet = (): ThemeSetContextType => {
  const context = useContext(ThemeSetContext)

  if (context === undefined) {
    throw new Error("useThemeSet must be used within a ThemeSetProvider")
  }

  return context
} 