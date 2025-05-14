import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Palette } from "lucide-react"
import { AccentColor, ThemeSet, useThemeSet } from "./theme-set-provider"
import { cn } from "@/lib/utils"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { useState } from "react"

type ThemeSetOption = {
  value: ThemeSet
  label: string
  colors: string[]
}

const themeSetOptions: ThemeSetOption[] = [
  {
    value: "default",
    label: "Monochrome",
    colors: ["#FFFFFF", "#000000"]
  },
  {
    value: "patagonia",
    label: "Patagonia",
    colors: ["#FDFBEE", "#57B4BA", "#015551", "#FE4F2D"]
  },
  {
    value: "redwood",
    label: "Redwood",
    colors: ["#FEF9E1", "#E5D0AC", "#A31D1D", "#6D2323"]
  }
]

type AccentOption = {
  value: AccentColor
  label: string
  color: string
}

const accentOptions: AccentOption[] = [
  {
    value: "indigo",
    label: "Indigo",
    color: "hsl(240, 70%, 50%)"
  },
  {
    value: "rose",
    label: "Rose",
    color: "hsl(350, 70%, 50%)"
  },
  {
    value: "forest",
    label: "Forest",
    color: "hsl(150, 70%, 35%)"
  },
  {
    value: "amber",
    label: "Amber",
    color: "hsl(40, 100%, 50%)"
  },
  {
    value: "teal",
    label: "Teal",
    color: "hsl(175, 70%, 40%)"
  }
]

interface ThemeSetToggleProps {
  className?: string
  compact?: boolean
}

export function ThemeSetToggle({ className, compact = false }: ThemeSetToggleProps) {
  const { themeSet, setThemeSet, accent, setAccent } = useThemeSet()
  const [showingAccent, setShowingAccent] = useState(false)

  return (
    <Popover onOpenChange={() => setShowingAccent(false)}>
      <Tooltip>
        <TooltipTrigger asChild>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              className={cn(
                "w-10 h-10 rounded-full",
                compact ? "border-0" : "",
                className
              )}
            >
              <Palette className="h-4 w-4" />
              <span className="sr-only">Toggle theme set</span>
            </Button>
          </PopoverTrigger>
        </TooltipTrigger>
        <TooltipContent side={compact ? "right" : "bottom"}>
          Theme Set
        </TooltipContent>
      </Tooltip>
      <PopoverContent className="w-56">
        {!showingAccent ? (
          <div className="grid gap-2">
            <div className="space-y-1">
              <h4 className="font-medium leading-none">Theme Set</h4>
              <p className="text-xs text-muted-foreground">
                Select a color palette for the application.
              </p>
            </div>
            <div className="grid gap-1">
              {themeSetOptions.map((option) => (
                <Button
                  key={option.value}
                  variant={option.value === themeSet ? "default" : "outline"}
                  className="justify-start gap-2"
                  onClick={() => {
                    setThemeSet(option.value);
                    if (option.value === "default") {
                      setShowingAccent(true);
                    }
                  }}
                >
                  <div className="flex gap-1 items-center">
                    {option.colors.map((color, i) => (
                      <div 
                        key={i}
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: color }}
                      />
                    ))}
                  </div>
                  <span>{option.label}</span>
                </Button>
              ))}
            </div>
          </div>
        ) : (
          <div className="grid gap-2">
            <div className="space-y-1">
              <div className="flex items-center justify-between">
                <h4 className="font-medium leading-none">Accent Color</h4>
                <Button 
                  variant="ghost" 
                  size="sm" 
                  className="h-8 px-2 text-xs"
                  onClick={() => setShowingAccent(false)}
                >
                  Back
                </Button>
              </div>
              <p className="text-xs text-muted-foreground">
                Choose an accent color for the monochrome theme.
              </p>
            </div>
            <div className="grid gap-1">
              {accentOptions.map((option) => (
                <Button
                  key={option.value}
                  variant={option.value === accent ? "default" : "outline"}
                  className="justify-start gap-2"
                  onClick={() => setAccent(option.value)}
                >
                  <div 
                    className="w-4 h-4 rounded-full"
                    style={{ backgroundColor: option.color }}
                  />
                  <span>{option.label}</span>
                </Button>
              ))}
            </div>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )
} 