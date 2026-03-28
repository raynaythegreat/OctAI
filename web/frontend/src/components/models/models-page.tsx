import {
  IconKey,
  IconLoader2,
  IconPlus,
  IconStar,
  IconTrash,
} from "@tabler/icons-react"
import { useCallback, useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  type ModelFormData,
  type ModelInfo,
  addImageModel,
  addVideoModel,
  deleteImageModel,
  deleteVideoModel,
  getImageModelsFiltered,
  getModels,
  getVideoModelsFiltered,
  rotateImageModelKey,
  rotateModelKey,
  rotateVideoModelKey,
  setDefaultModel,
  testImageModelKey,
  testVideoModelKey,
} from "@/api/models"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Sheet,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { Switch } from "@/components/ui/switch"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

import { AddModelSheet } from "./add-model-sheet"
import { DeleteModelDialog } from "./delete-model-dialog"
import { EditModelSheet } from "./edit-model-sheet"
import { getProviderKey, getProviderLabel } from "./provider-label"
import { ProviderSection } from "./provider-section"

const PROVIDER_PRIORITY: Record<string, number> = {
  volcengine: 0,
  openai: 1,
  gemini: 2,
  anthropic: 3,
  zhipu: 4,
  deepseek: 5,
  openrouter: 6,
  qwen: 7,
  moonshot: 8,
  groq: 9,
  "github-copilot": 10,
  antigravity: 11,
  nvidia: 12,
  cerebras: 13,
  shengsuanyun: 14,
  ollama: 15,
  vllm: 16,
  mistral: 17,
  avian: 18,
  mimo: 19,
}

interface ProviderGroup {
  key: string
  label: string
  models: ModelInfo[]
  hasDefault: boolean
  configuredCount: number
}

// ─── Rotate Key Dialog ────────────────────────────────────────────────────────

interface RotateKeyDialogProps {
  open: boolean
  onClose: () => void
  onConfirm: (newKey: string) => Promise<void>
  modelName: string
}

function RotateKeySheet({
  open,
  onClose,
  onConfirm,
  modelName,
}: RotateKeyDialogProps) {
  const { t } = useTranslation()
  const [newKey, setNewKey] = useState("")
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open) setNewKey("")
  }, [open])

  const handleConfirm = async () => {
    if (!newKey.trim()) return
    setSaving(true)
    try {
      await onConfirm(newKey.trim())
      onClose()
    } finally {
      setSaving(false)
    }
  }

  return (
    <Sheet open={open} onOpenChange={(v) => !v && onClose()}>
      <SheetContent
        side="right"
        className="flex flex-col gap-0 p-0 data-[side=right]:!w-full data-[side=right]:sm:!w-[420px] data-[side=right]:sm:!max-w-[420px]"
      >
        <SheetHeader className="border-b-muted border-b px-6 py-5">
          <SheetTitle className="text-base">
            {t("models.rotateKey.title", { name: modelName })}
          </SheetTitle>
        </SheetHeader>

        <div className="flex-1 px-6 py-5">
          <div className="space-y-2">
            <Label htmlFor="new-api-key">{t("models.rotateKey.label")}</Label>
            <Input
              id="new-api-key"
              type="password"
              value={newKey}
              onChange={(e) => setNewKey(e.target.value)}
              placeholder={t("models.rotateKey.placeholder")}
              onKeyDown={(e) => {
                if (e.key === "Enter") void handleConfirm()
              }}
            />
          </div>
        </div>

        <SheetFooter className="border-t-muted border-t px-6 py-4">
          <Button variant="ghost" onClick={onClose} disabled={saving}>
            {t("common.cancel")}
          </Button>
          <Button onClick={handleConfirm} disabled={saving || !newKey.trim()}>
            {saving && <IconLoader2 className="size-4 animate-spin" />}
            {t("models.rotateKey.confirm")}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}

// ─── Add Media Model Sheet ─────────────────────────────────────────────────────

interface AddMediaModelSheetProps {
  open: boolean
  onClose: () => void
  onSaved: () => void
  type: "image" | "video"
}

function AddMediaModelSheet({
  open,
  onClose,
  onSaved,
  type,
}: AddMediaModelSheetProps) {
  const { t } = useTranslation()
  const [form, setForm] = useState<ModelFormData>({
    model_name: "",
    model: "",
    api_base: "",
    api_key: "",
  })
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => {
    if (open) {
      setForm({ model_name: "", model: "", api_base: "", api_key: "" })
      setError("")
    }
  }, [open])

  const handleSave = async () => {
    if (!form.model_name.trim() || !form.model.trim()) {
      setError(t("models.add.errorRequired"))
      return
    }
    setSaving(true)
    setError("")
    try {
      const payload: ModelFormData = {
        model_name: form.model_name.trim(),
        model: form.model.trim(),
        api_base: form.api_base?.trim() || undefined,
        api_key: form.api_key?.trim() || undefined,
      }
      if (type === "image") {
        await addImageModel(payload)
      } else {
        await addVideoModel(payload)
      }
      onSaved()
      onClose()
    } catch (e) {
      setError(e instanceof Error ? e.message : t("models.add.saveError"))
    } finally {
      setSaving(false)
    }
  }

  const titleKey = type === "image" ? "models.addImage.title" : "models.addVideo.title"

  return (
    <Sheet open={open} onOpenChange={(v) => !v && onClose()}>
      <SheetContent
        side="right"
        className="flex flex-col gap-0 p-0 data-[side=right]:!w-full data-[side=right]:sm:!w-[480px] data-[side=right]:sm:!max-w-[480px]"
      >
        <SheetHeader className="border-b-muted border-b px-6 py-5">
          <SheetTitle className="text-base">
            {t(titleKey, {
              defaultValue:
                type === "image" ? "Add Image Model" : "Add Video Model",
            })}
          </SheetTitle>
        </SheetHeader>

        <div className="min-h-0 flex-1 overflow-y-auto">
          <div className="space-y-4 px-6 py-5">
            <div className="space-y-2">
              <Label htmlFor="media-model-name">
                {t("models.add.modelName")}
              </Label>
              <Input
                id="media-model-name"
                value={form.model_name}
                onChange={(e) =>
                  setForm((f) => ({ ...f, model_name: e.target.value }))
                }
                placeholder={t("models.add.modelNamePlaceholder")}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="media-model-id">{t("models.add.modelId")}</Label>
              <Input
                id="media-model-id"
                value={form.model}
                onChange={(e) =>
                  setForm((f) => ({ ...f, model: e.target.value }))
                }
                placeholder={
                  type === "image" ? "openai/dall-e-3" : "runway/gen-3"
                }
                className="font-mono text-sm"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="media-api-base">
                {t("models.field.apiBase")}
              </Label>
              <Input
                id="media-api-base"
                value={form.api_base ?? ""}
                onChange={(e) =>
                  setForm((f) => ({ ...f, api_base: e.target.value }))
                }
                placeholder="https://api.example.com/v1"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="media-api-key">{t("models.field.apiKey")}</Label>
              <Input
                id="media-api-key"
                type="password"
                value={form.api_key ?? ""}
                onChange={(e) =>
                  setForm((f) => ({ ...f, api_key: e.target.value }))
                }
                placeholder={t("models.field.apiKeyPlaceholder")}
              />
            </div>

            {error && (
              <p className="text-destructive bg-destructive/10 rounded-md px-3 py-2 text-sm">
                {error}
              </p>
            )}
          </div>
        </div>

        <SheetFooter className="border-t-muted border-t px-6 py-4">
          <Button variant="ghost" onClick={onClose} disabled={saving}>
            {t("common.cancel")}
          </Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving && <IconLoader2 className="size-4 animate-spin" />}
            {t("models.add.confirm")}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}

// ─── Media Model Card ─────────────────────────────────────────────────────────

interface MediaModelCardProps {
  model: ModelInfo
  type: "image" | "video"
  onDeleted: () => void
  onRotateKey: (model: ModelInfo) => void
}

function MediaModelCard({
  model,
  type,
  onDeleted,
  onRotateKey,
}: MediaModelCardProps) {
  const { t } = useTranslation()
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<{
    success: boolean
    error?: string
    models?: string[]
    note?: string
  } | null>(null)

  const handleTest = async () => {
    setTesting(true)
    setTestResult(null)
    try {
      const result =
        type === "image"
          ? await testImageModelKey(model.index)
          : await testVideoModelKey(model.index)
      setTestResult(result)
    } catch (e) {
      setTestResult({
        success: false,
        error: e instanceof Error ? e.message : "Test failed",
      })
    } finally {
      setTesting(false)
    }
  }

  const handleDelete = async () => {
    try {
      if (type === "image") {
        await deleteImageModel(model.index)
      } else {
        await deleteVideoModel(model.index)
      }
      toast.success(
        t("models.delete.success", {
          name: model.model_name,
          defaultValue: `${model.model_name} deleted`,
        }),
      )
      onDeleted()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Delete failed")
    }
  }

  return (
    <div className="bg-card border-border/60 rounded-xl border p-4 shadow-sm">
      <div className="mb-2 flex items-start justify-between gap-2">
        <div className="min-w-0">
          <h3 className="truncate text-sm font-semibold">{model.model_name}</h3>
          <p className="text-muted-foreground mt-0.5 truncate font-mono text-xs">
            {model.model}
          </p>
        </div>
        <span
          className={`shrink-0 text-xs ${model.configured ? "text-emerald-600 dark:text-emerald-400" : "text-muted-foreground"}`}
        >
          {model.configured
            ? t("credentials.apikey.configured")
            : t("credentials.apikey.notConfigured")}
        </span>
      </div>

      {testResult && (
        <div
          className={`mb-2 rounded-md px-2.5 py-1.5 text-xs ${testResult.success ? "bg-emerald-50 text-emerald-700 dark:bg-emerald-950/30 dark:text-emerald-400" : "bg-destructive/10 text-destructive"}`}
        >
          {testResult.success ? (
            <>
              <span className="font-medium">
                {t("models.test.success", { defaultValue: "Connected" })}
              </span>
              {testResult.models && testResult.models.length > 0 && (
                <span className="ml-1 opacity-75">
                  &middot;{" "}
                  {t("models.test.modelsCount", {
                    count: testResult.models.length,
                    defaultValue: `${testResult.models.length} models available`,
                  })}
                </span>
              )}
              {testResult.note && (
                <span className="ml-1 opacity-75">&middot; {testResult.note}</span>
              )}
            </>
          ) : (
            testResult.error ?? t("models.test.failed", { defaultValue: "Failed" })
          )}
        </div>
      )}

      <div className="flex items-center gap-1.5">
        <Button
          size="sm"
          variant="outline"
          className="h-7 text-xs"
          onClick={handleTest}
          disabled={testing}
        >
          {testing ? (
            <IconLoader2 className="size-3 animate-spin" />
          ) : null}
          {t("models.action.test", { defaultValue: "Test" })}
        </Button>

        <Button
          size="icon-sm"
          variant="ghost"
          onClick={() => onRotateKey(model)}
          title={t("models.rotateKey.title", {
            name: model.model_name,
            defaultValue: "Rotate API key",
          })}
          className="h-7 w-7"
        >
          <IconKey className="size-3.5" />
        </Button>

        <Button
          size="icon-sm"
          variant="ghost"
          onClick={handleDelete}
          title={t("models.action.delete")}
          className="text-muted-foreground hover:text-destructive hover:bg-destructive/10 ml-auto h-7 w-7"
        >
          <IconTrash className="size-3.5" />
        </Button>
      </div>
    </div>
  )
}

// ─── Main Page ────────────────────────────────────────────────────────────────

export function ModelsPage() {
  const { t } = useTranslation()
  const [models, setModels] = useState<ModelInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [fetchError, setFetchError] = useState("")
  const [configuredOnly, setConfiguredOnly] = useState(true)

  const [editingModel, setEditingModel] = useState<ModelInfo | null>(null)
  const [deletingModel, setDeletingModel] = useState<ModelInfo | null>(null)
  const [addOpen, setAddOpen] = useState(false)
  const [settingDefaultIndex, setSettingDefaultIndex] = useState<number | null>(
    null,
  )

  // Rotate key state
  const [rotatingModel, setRotatingModel] = useState<{
    model: ModelInfo
    type: "text" | "image" | "video"
  } | null>(null)

  const [activeTab, setActiveTab] = useState("text")
  const [imageModels, setImageModels] = useState<ModelInfo[]>([])
  const [videoModels, setVideoModels] = useState<ModelInfo[]>([])
  const [imageModelsLoading, setImageModelsLoading] = useState(false)
  const [videoModelsLoading, setVideoModelsLoading] = useState(false)

  // Add media model sheets
  const [addImageOpen, setAddImageOpen] = useState(false)
  const [addVideoOpen, setAddVideoOpen] = useState(false)

  const fetchModels = useCallback(async () => {
    try {
      const data = await getModels({ configured_only: configuredOnly })
      const sorted = [...data.models].sort((a, b) => {
        if (a.is_default && !b.is_default) return -1
        if (!a.is_default && b.is_default) return 1
        if (a.configured && !b.configured) return -1
        if (!a.configured && b.configured) return 1
        return a.model_name.localeCompare(b.model_name)
      })
      setModels(sorted)
      setFetchError("")
    } catch (e) {
      setFetchError(e instanceof Error ? e.message : t("models.loadError"))
    } finally {
      setLoading(false)
    }
  }, [t, configuredOnly])

  const fetchImageModels = useCallback(async () => {
    setImageModelsLoading(true)
    try {
      const data = await getImageModelsFiltered({
        configured_only: configuredOnly,
      })
      setImageModels(data.models)
    } catch {
      // silent
    } finally {
      setImageModelsLoading(false)
    }
  }, [configuredOnly])

  const fetchVideoModels = useCallback(async () => {
    setVideoModelsLoading(true)
    try {
      const data = await getVideoModelsFiltered({
        configured_only: configuredOnly,
      })
      setVideoModels(data.models)
    } catch {
      // silent
    } finally {
      setVideoModelsLoading(false)
    }
  }, [configuredOnly])

  useEffect(() => {
    void fetchModels()
  }, [fetchModels])

  useEffect(() => {
    if (activeTab === "image") {
      void fetchImageModels()
    } else if (activeTab === "video") {
      void fetchVideoModels()
    }
  }, [activeTab, fetchImageModels, fetchVideoModels])

  const handleSetDefault = async (model: ModelInfo) => {
    if (model.is_default) return

    setSettingDefaultIndex(model.index)
    try {
      await setDefaultModel(model.model_name)
      await fetchModels()
    } catch {
      // ignore
    } finally {
      setSettingDefaultIndex(null)
    }
  }

  const handleRotateKey = async (newKey: string) => {
    if (!rotatingModel) return
    const { model, type } = rotatingModel
    try {
      if (type === "text") {
        await rotateModelKey(model.index, newKey)
      } else if (type === "image") {
        await rotateImageModelKey(model.index, newKey)
      } else {
        await rotateVideoModelKey(model.index, newKey)
      }
      toast.success(
        t("models.rotateKey.success", {
          name: model.model_name,
          defaultValue: `API key rotated for ${model.model_name}`,
        }),
      )
      if (type === "text") void fetchModels()
      else if (type === "image") void fetchImageModels()
      else void fetchVideoModels()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to rotate key")
      throw e
    }
  }

  const grouped: Record<string, { label: string; models: ModelInfo[] }> = {}
  for (const model of models) {
    const providerKey = getProviderKey(model.model)
    if (!grouped[providerKey]) {
      grouped[providerKey] = {
        label: getProviderLabel(model.model),
        models: [],
      }
    }
    grouped[providerKey].models.push(model)
  }

  const providerGroups: ProviderGroup[] = Object.entries(grouped)
    .map(([key, group]) => {
      const configuredCount = group.models.filter(
        (model) => model.configured,
      ).length
      return {
        key,
        label: group.label,
        models: group.models,
        hasDefault: group.models.some((model) => model.is_default),
        configuredCount,
      }
    })
    .sort((a, b) => {
      if (a.hasDefault && !b.hasDefault) return -1
      if (!a.hasDefault && b.hasDefault) return 1

      if (a.configuredCount !== b.configuredCount) {
        return b.configuredCount - a.configuredCount
      }

      const aPriority = PROVIDER_PRIORITY[a.key] ?? Number.MAX_SAFE_INTEGER
      const bPriority = PROVIDER_PRIORITY[b.key] ?? Number.MAX_SAFE_INTEGER
      if (aPriority !== bPriority) {
        return aPriority - bPriority
      }

      return a.label.localeCompare(b.label)
    })

  const defaultModel = models.find((model) => model.is_default)

  // Configured-only toggle (shared between tabs)
  const ConfiguredOnlyToggle = (
    <div className="flex items-center gap-2">
      <Switch
        id="configured-only"
        size="sm"
        checked={configuredOnly}
        onCheckedChange={setConfiguredOnly}
      />
      <Label
        htmlFor="configured-only"
        className="text-muted-foreground cursor-pointer text-xs"
      >
        {t("models.filter.configuredOnly", {
          defaultValue: "Show configured only",
        })}
      </Label>
    </div>
  )

  return (
    <div className="flex h-full flex-col">
      <PageHeader title={t("navigation.models")}>
        {activeTab === "text" && (
          <div className="flex items-center gap-3">
            {ConfiguredOnlyToggle}
            <Button size="sm" variant="outline" onClick={() => setAddOpen(true)}>
              <IconPlus className="size-4" />
              {t("models.add.button")}
            </Button>
          </div>
        )}
        {activeTab === "image" && (
          <div className="flex items-center gap-3">
            {ConfiguredOnlyToggle}
            <Button
              size="sm"
              variant="outline"
              onClick={() => setAddImageOpen(true)}
            >
              <IconPlus className="size-4" />
              {t("models.add.button")}
            </Button>
          </div>
        )}
        {activeTab === "video" && (
          <div className="flex items-center gap-3">
            {ConfiguredOnlyToggle}
            <Button
              size="sm"
              variant="outline"
              onClick={() => setAddVideoOpen(true)}
            >
              <IconPlus className="size-4" />
              {t("models.add.button")}
            </Button>
          </div>
        )}
      </PageHeader>

      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="flex flex-1 flex-col overflow-hidden"
      >
        <TabsList className="mx-4 mt-4 w-fit md:mx-6">
          <TabsTrigger value="text">{t("models.tabs.text")}</TabsTrigger>
          <TabsTrigger value="image">{t("models.tabs.image")}</TabsTrigger>
          <TabsTrigger value="video">{t("models.tabs.video")}</TabsTrigger>
        </TabsList>

        <TabsContent value="text" className="flex-1 overflow-auto">
          <div className="min-h-0 flex-1 px-4 sm:px-6">
            <div className="pt-2">
              {!defaultModel && (
                <div className="text-muted-foreground flex items-center gap-1.5 text-sm">
                  <span>{t("models.noDefaultHintPrefix")}</span>
                  <IconStar className="size-3.5 shrink-0" />
                  <span>{t("models.noDefaultHintSuffix")}</span>
                </div>
              )}
              <p className="text-muted-foreground mt-1 text-sm">
                {t("models.description")}
              </p>
            </div>

            {loading && (
              <div className="flex items-center justify-center py-20">
                <IconLoader2 className="text-muted-foreground size-6 animate-spin" />
              </div>
            )}

            {fetchError && (
              <div className="text-destructive bg-destructive/10 rounded-lg px-4 py-3 text-sm">
                {fetchError}
              </div>
            )}

            {!loading && !fetchError && (
              <div className="pb-8">
                {providerGroups.map((providerGroup) => (
                  <ProviderSection
                    key={providerGroup.key}
                    provider={providerGroup.label}
                    providerKey={providerGroup.key}
                    models={providerGroup.models}
                    onEdit={setEditingModel}
                    onSetDefault={handleSetDefault}
                    onDelete={setDeletingModel}
                    onRotateKey={(model) =>
                      setRotatingModel({ model, type: "text" })
                    }
                    settingDefaultIndex={settingDefaultIndex}
                  />
                ))}
              </div>
            )}
          </div>
        </TabsContent>

        <TabsContent value="image" className="flex-1 overflow-auto p-4 md:p-6">
          {imageModelsLoading ? (
            <div className="flex items-center justify-center py-20">
              <IconLoader2 className="text-muted-foreground size-6 animate-spin" />
            </div>
          ) : imageModels.length === 0 ? (
            <p className="text-muted-foreground text-sm">
              {t("models.noModels")}
            </p>
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {imageModels.map((model) => (
                <MediaModelCard
                  key={model.index}
                  model={model}
                  type="image"
                  onDeleted={fetchImageModels}
                  onRotateKey={(m) => setRotatingModel({ model: m, type: "image" })}
                />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="video" className="flex-1 overflow-auto p-4 md:p-6">
          {videoModelsLoading ? (
            <div className="flex items-center justify-center py-20">
              <IconLoader2 className="text-muted-foreground size-6 animate-spin" />
            </div>
          ) : videoModels.length === 0 ? (
            <p className="text-muted-foreground text-sm">
              {t("models.noModels")}
            </p>
          ) : (
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {videoModels.map((model) => (
                <MediaModelCard
                  key={model.index}
                  model={model}
                  type="video"
                  onDeleted={fetchVideoModels}
                  onRotateKey={(m) => setRotatingModel({ model: m, type: "video" })}
                />
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>

      <EditModelSheet
        model={editingModel}
        open={editingModel !== null}
        onClose={() => setEditingModel(null)}
        onSaved={fetchModels}
      />

      <AddModelSheet
        open={addOpen}
        onClose={() => setAddOpen(false)}
        onSaved={fetchModels}
        existingModelNames={models.map((model) => model.model_name)}
      />

      <DeleteModelDialog
        model={deletingModel}
        onClose={() => setDeletingModel(null)}
        onDeleted={fetchModels}
      />

      <AddMediaModelSheet
        open={addImageOpen}
        onClose={() => setAddImageOpen(false)}
        onSaved={fetchImageModels}
        type="image"
      />

      <AddMediaModelSheet
        open={addVideoOpen}
        onClose={() => setAddVideoOpen(false)}
        onSaved={fetchVideoModels}
        type="video"
      />

      {rotatingModel && (
        <RotateKeySheet
          open={rotatingModel !== null}
          onClose={() => setRotatingModel(null)}
          onConfirm={handleRotateKey}
          modelName={rotatingModel.model.model_name}
        />
      )}
    </div>
  )
}
