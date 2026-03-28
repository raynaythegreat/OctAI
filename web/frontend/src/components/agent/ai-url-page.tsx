import {
  IconCheck,
  IconLink,
  IconLoader2,
  IconPlug,
  IconPuzzle,
  IconSparkles,
  IconTools,
} from "@tabler/icons-react"
import * as React from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import { type DiscoveredItem, analyzeURL, integrateItems } from "@/api/scanner"
import { PageHeader } from "@/components/page-header"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { ScrollArea } from "@/components/ui/scroll-area"

const TYPE_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  mcp_server: IconPlug,
  skill: IconSparkles,
  tool: IconTools,
  plugin: IconPuzzle,
  connection: IconLink,
}

export function AiUrlPage() {
  const { t } = useTranslation()
  const [url, setUrl] = React.useState("")
  const [scanning, setScanning] = React.useState(false)
  const [items, setItems] = React.useState<DiscoveredItem[] | null>(null)
  const [selected, setSelected] = React.useState<Set<number>>(new Set())
  const [integrating, setIntegrating] = React.useState(false)

  const handleScan = async () => {
    const trimmed = url.trim()
    if (!trimmed) {
      toast.error(t("aiUrl.errors.emptyUrl"))
      return
    }
    setScanning(true)
    setItems(null)
    setSelected(new Set())
    try {
      const result = await analyzeURL(trimmed)
      setItems(result.items)
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : t("aiUrl.errors.analyzeFailed"),
      )
    } finally {
      setScanning(false)
    }
  }

  const toggleItem = (index: number) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(index)) {
        next.delete(index)
      } else {
        next.add(index)
      }
      return next
    })
  }

  const toggleAll = () => {
    if (!items) return
    if (selected.size === items.length) {
      setSelected(new Set())
    } else {
      setSelected(new Set(items.map((_, i) => i)))
    }
  }

  const handleIntegrate = async () => {
    if (!items || selected.size === 0) return
    const toIntegrate = [...selected].map((i) => items[i])
    setIntegrating(true)
    try {
      const results = await integrateItems(toIntegrate)
      const succeeded = results.filter((r) => r.success).length
      const failed = results.filter((r) => !r.success)
      if (succeeded > 0) {
        toast.success(
          t("aiUrl.integrateSuccess", { count: succeeded }),
        )
      }
      for (const f of failed) {
        toast.error(`${f.name}: ${f.error ?? t("aiUrl.errors.integrateFailed")}`)
      }
      if (succeeded > 0) {
        setSelected(new Set())
        // Notify chat composer and other listeners to refresh their skill lists
        window.dispatchEvent(new CustomEvent("skills-updated"))
      }
    } catch (err) {
      toast.error(
        err instanceof Error ? err.message : t("aiUrl.errors.integrateFailed"),
      )
    } finally {
      setIntegrating(false)
    }
  }

  const allSelected = items !== null && items.length > 0 && selected.size === items.length

  return (
    <div className="flex h-full flex-col">
      <PageHeader title={t("aiUrl.title")} />
      <div className="flex min-h-0 flex-1 flex-col gap-4 p-4 md:p-6">
        <p className="text-muted-foreground shrink-0 text-sm">
          {t("aiUrl.description")}
        </p>

        {/* URL Input */}
        <div className="flex shrink-0 gap-2">
          <Input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder={t("aiUrl.inputPlaceholder")}
            className="flex-1"
            onKeyDown={(e) => {
              if (e.key === "Enter") void handleScan()
            }}
            disabled={scanning}
          />
          <Button onClick={handleScan} disabled={scanning || !url.trim()}>
            {scanning ? (
              <IconLoader2 className="size-4 animate-spin" />
            ) : (
              t("aiUrl.scanButton")
            )}
          </Button>
        </div>

        {/* Scanning state */}
        {scanning && (
          <div className="text-muted-foreground flex shrink-0 items-center gap-2 text-sm">
            <IconLoader2 className="size-4 animate-spin" />
            {t("aiUrl.scanning")}
          </div>
        )}

        {/* Results */}
        {items !== null && !scanning && (
          <div className="flex min-h-0 flex-1 flex-col gap-3">
            {items.length === 0 ? (
              <div className="flex flex-1 items-center justify-center text-muted-foreground">
                <p>{t("aiUrl.noResults")}</p>
              </div>
            ) : (
              <>
                {/* Header row */}
                <div className="flex shrink-0 items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Checkbox
                      checked={allSelected}
                      onCheckedChange={toggleAll}
                      id="select-all"
                    />
                    <label htmlFor="select-all" className="cursor-pointer text-sm">
                      {t("aiUrl.selectAll")} •{" "}
                      <span className="text-muted-foreground">
                        {t("aiUrl.resultsTitle", { count: items.length })}
                      </span>
                    </label>
                  </div>
                  {selected.size > 0 && (
                    <Button
                      size="sm"
                      onClick={handleIntegrate}
                      disabled={integrating}
                    >
                      {integrating ? (
                        <IconLoader2 className="mr-2 size-4 animate-spin" />
                      ) : (
                        <IconCheck className="mr-2 size-4" />
                      )}
                      {t("aiUrl.integrateButton", { count: selected.size })}
                    </Button>
                  )}
                </div>

                {/* Item list — min-h-0 is required for flex children to scroll */}
                <ScrollArea className="min-h-0 flex-1">
                  <div className="flex flex-col gap-2 pb-2 pr-4">
                    {items.map((item, i) => {
                      const Icon = TYPE_ICONS[item.type] ?? IconLink
                      const isSelected = selected.has(i)
                      return (
                        <div
                          key={i}
                          className={`flex cursor-pointer items-start gap-3 rounded-lg border p-3 transition-colors ${
                            isSelected
                              ? "border-primary/50 bg-primary/5"
                              : "hover:bg-muted/50"
                          }`}
                          onClick={() => toggleItem(i)}
                        >
                          <Checkbox
                            checked={isSelected}
                            onCheckedChange={() => toggleItem(i)}
                            onClick={(e) => e.stopPropagation()}
                            className="mt-0.5"
                          />
                          <Icon className="mt-0.5 size-5 shrink-0 text-muted-foreground" />
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center gap-2">
                              <span className="text-sm font-medium">{item.name}</span>
                              <Badge variant="secondary" className="text-xs">
                                {t(`aiUrl.types.${item.type}`, { defaultValue: item.type })}
                              </Badge>
                            </div>
                            {item.description && (
                              <p className="text-muted-foreground mt-0.5 line-clamp-2 text-xs">
                                {item.description}
                              </p>
                            )}
                          </div>
                        </div>
                      )
                    })}
                  </div>
                </ScrollArea>
              </>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
