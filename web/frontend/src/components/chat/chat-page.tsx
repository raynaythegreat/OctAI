import { IconBrain, IconBug, IconHammer, IconMessageCircle, IconPlus, IconSparkles, IconSearch } from "@tabler/icons-react"
import { useCallback, useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"

import { getSessionHistory } from "@/api/sessions"
import { AssistantMessage } from "@/components/chat/assistant-message"
import { ChannelSelector } from "@/components/chat/channel-selector"
import { ChatComposer } from "@/components/chat/chat-composer"
import { ChatEmptyState } from "@/components/chat/chat-empty-state"
import { ModelSelector } from "@/components/chat/model-selector"
import { SessionHistoryMenu } from "@/components/chat/session-history-menu"
import { TypingIndicator } from "@/components/chat/typing-indicator"
import { UserMessage } from "@/components/chat/user-message"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { useChatModels } from "@/hooks/use-chat-models"
import { useGateway } from "@/hooks/use-gateway"
import { usePicoChat } from "@/hooks/use-pico-chat"
import { useSessionHistory } from "@/hooks/use-session-history"

export function ChatPage() {
  const { t } = useTranslation()
  const scrollRef = useRef<HTMLDivElement>(null)
  const [isAtBottom, setIsAtBottom] = useState(true)
  const [hasScrolled, setHasScrolled] = useState(false)
  const [input, setInput] = useState("")
  const [activeChannel, setActiveChannel] = useState("pico")
  const [viewingSessionId, setViewingSessionId] = useState<string | null>(null)
  const [viewingMessages, setViewingMessages] = useState<{ role: string; content: string }[] | null>(null)
  const [chatMode, setChatMode] = useState<"chat" | "plan" | "build">("build")

  const readOnly = activeChannel !== "pico"

  const {
    messages,
    connectionState,
    isTyping,
    activeSessionId,
    sendMessage,
    switchSession,
    newChat,
  } = usePicoChat()

  const { state: gwState } = useGateway()
  const isGatewayRunning = gwState === "running"
  const isChatConnected = connectionState === "connected"

  const {
    defaultModelName,
    hasConfiguredModels,
    apiKeyModels,
    oauthModels,
    localModels,
    handleSetDefault,
    isAutoMode,
    toggleAutoMode,
  } = useChatModels({ isConnected: isGatewayRunning })
  const canSend = isChatConnected && Boolean(defaultModelName)

  const {
    sessions,
    hasMore,
    loadError,
    loadErrorMessage,
    observerRef,
    loadSessions,
    handleDeleteSession,
  } = useSessionHistory({
    activeSessionId,
    onDeletedActiveSession: newChat,
    channel: activeChannel !== "pico" ? activeChannel : undefined,
  })

  const loadReadOnlySession = useCallback(
    async (id: string) => {
      try {
        const detail = await getSessionHistory(id, activeChannel)
        setViewingSessionId(id)
        setViewingMessages(detail.messages)
      } catch (err) {
        console.error("Failed to load session:", err)
      }
    },
    [activeChannel],
  )

  const syncScrollState = (element: HTMLDivElement) => {
    const { scrollTop, scrollHeight, clientHeight } = element
    setHasScrolled(scrollTop > 0)
    setIsAtBottom(scrollHeight - scrollTop <= clientHeight + 10)
  }

  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    syncScrollState(e.currentTarget)
  }

  useEffect(() => {
    if (scrollRef.current) {
      if (isAtBottom) {
        scrollRef.current.scrollTop = scrollRef.current.scrollHeight
      }
      syncScrollState(scrollRef.current)
    }
  }, [messages, isTyping, isAtBottom])

  const applyModePrefix = (content: string) => {
    if (chatMode === "plan") {
      return `Before taking any action, write a clear numbered plan of what you will do. Then execute it step by step.\n\n${content}`
    }
    if (chatMode === "chat") {
      return `You are in conversational mode. Focus on discussion, research, and answering questions. Do not make file changes or run commands unless explicitly asked.\n\n${content}`
    }
    return content
  }

  const cycleChatMode = () => {
    setChatMode((m) => (m === "build" ? "chat" : m === "chat" ? "plan" : "build"))
  }

  const handleSend = () => {
    if (!input.trim() || !canSend) return
    if (sendMessage(applyModePrefix(input.trim()))) {
      setInput("")
    }
  }

  const handleSendWithAttachments = async (
    content: string,
    attachments: { file: File; dataUrl?: string }[],
  ) => {
    if (!canSend) return
    let fullContent = content
    for (const att of attachments) {
      if (att.dataUrl && att.dataUrl.startsWith("data:image/")) {
        // Note the attached image by name; vision-capable models receive context via the UI
        fullContent += `\n\n[Attached image: ${att.file.name}]`
      } else {
        // For text files, inline the content
        try {
          const text = await att.file.text()
          fullContent += `\n\n\`\`\`\n${att.file.name}:\n${text}\`\`\``
        } catch {
          fullContent += `\n\n[Attached file: ${att.file.name}]`
        }
      }
    }
    if (sendMessage(applyModePrefix(fullContent))) {
      setInput("")
    }
  }

  return (
    <div className="bg-background/95 flex h-full flex-col">
      <PageHeader
        title={t("navigation.chat")}
        className={`transition-shadow ${
          hasScrolled ? "shadow-sm" : "shadow-none"
        }`}
        titleExtra={
          <div className="flex items-center gap-2">
            {hasConfiguredModels && (
              <ModelSelector
                defaultModelName={defaultModelName}
                apiKeyModels={apiKeyModels}
                oauthModels={oauthModels}
                localModels={localModels}
                onValueChange={handleSetDefault}
                isAutoMode={isAutoMode}
                toggleAutoMode={toggleAutoMode}
              />
            )}
            <ChannelSelector
              activeChannel={activeChannel}
              onChannelChange={(ch) => {
                setActiveChannel(ch)
                setViewingSessionId(null)
                setViewingMessages(null)
              }}
            />
          </div>
        }
      >
        <Button
          variant="secondary"
          size="sm"
          onClick={
            readOnly
              ? () => {
                  setViewingSessionId(null)
                  setViewingMessages(null)
                }
              : newChat
          }
          className="h-9 gap-2"
        >
          <IconPlus className="size-4" />
          <span className="hidden sm:inline">{t("chat.newChat")}</span>
        </Button>

        <SessionHistoryMenu
          sessions={sessions}
          activeSessionId={readOnly ? (viewingSessionId ?? "") : activeSessionId}
          hasMore={hasMore}
          loadError={loadError}
          loadErrorMessage={loadErrorMessage}
          observerRef={observerRef}
          onOpenChange={(open) => {
            if (open) {
              void loadSessions(true)
            }
          }}
          onSwitchSession={readOnly ? loadReadOnlySession : switchSession}
          onDeleteSession={handleDeleteSession}
        />
      </PageHeader>

      <div
        ref={scrollRef}
        onScroll={handleScroll}
        className="min-h-0 flex-1 overflow-y-auto px-4 py-6 md:px-8 lg:px-24 xl:px-48"
      >
        <div className="mx-auto flex w-full max-w-250 flex-col gap-8 pb-8">
          {!readOnly && messages.length === 0 && !isTyping && (
            <>
              <ChatEmptyState
                hasConfiguredModels={hasConfiguredModels}
                defaultModelName={defaultModelName}
                isConnected={isGatewayRunning}
              />
              {canSend && (
                <div className="flex flex-wrap justify-center gap-2 px-4">
                  <button
                    type="button"
                    className="flex items-center gap-2 rounded-full border border-violet-500/30 bg-violet-500/5 px-4 py-2 text-sm text-violet-400 transition-colors hover:bg-violet-500/10"
                    onClick={() => { setChatMode("chat"); setInput("Research: ") }}
                  >
                    <IconSearch className="size-3.5" />
                    Research a topic
                  </button>
                  <button
                    type="button"
                    className="flex items-center gap-2 rounded-full border border-amber-500/30 bg-amber-500/5 px-4 py-2 text-sm text-amber-400 transition-colors hover:bg-amber-500/10"
                    onClick={() => { setChatMode("plan"); setInput("") }}
                  >
                    <IconBrain className="size-3.5" />
                    Write a plan
                  </button>
                  <button
                    type="button"
                    className="flex items-center gap-2 rounded-full border border-emerald-500/30 bg-emerald-500/5 px-4 py-2 text-sm text-emerald-400 transition-colors hover:bg-emerald-500/10"
                    onClick={() => { setChatMode("build"); setInput("") }}
                  >
                    <IconHammer className="size-3.5" />
                    Build something
                  </button>
                  <button
                    type="button"
                    className="flex items-center gap-2 rounded-full border border-violet-500/30 bg-violet-500/5 px-4 py-2 text-sm text-violet-400 transition-colors hover:bg-violet-500/10"
                    onClick={() => setInput("/use brainstorming ")}
                  >
                    <IconSparkles className="size-3.5" />
                    Brainstorm ideas
                  </button>
                  <button
                    type="button"
                    className="flex items-center gap-2 rounded-full border border-rose-500/30 bg-rose-500/5 px-4 py-2 text-sm text-rose-400 transition-colors hover:bg-rose-500/10"
                    onClick={() => setInput("/use systematic-debugging ")}
                  >
                    <IconBug className="size-3.5" />
                    Debug an issue
                  </button>
                  <button
                    type="button"
                    className="flex items-center gap-2 rounded-full border border-violet-500/30 bg-violet-500/5 px-4 py-2 text-sm text-violet-400 transition-colors hover:bg-violet-500/10"
                    onClick={() => { setChatMode("chat"); setInput("") }}
                  >
                    <IconMessageCircle className="size-3.5" />
                    Just chat
                  </button>
                </div>
              )}
            </>
          )}

          {!readOnly &&
            messages.map((msg) => (
              <div key={msg.id} className="flex w-full">
                {msg.role === "assistant" ? (
                  <AssistantMessage
                    content={msg.content}
                    timestamp={msg.timestamp}
                    meta={msg.meta}
                  />
                ) : (
                  <UserMessage content={msg.content} />
                )}
              </div>
            ))}

          {!readOnly && isTyping && <TypingIndicator />}

          {readOnly && !viewingMessages && (
            <div className="flex h-full items-center justify-center text-muted-foreground">
              <p>{t("chat.channel.noHistory", { channel: activeChannel.toUpperCase() })}</p>
            </div>
          )}

          {readOnly &&
            viewingMessages &&
            viewingMessages.map((msg, i) => (
              <div key={i} className="flex w-full">
                {msg.role === "assistant" ? (
                  <AssistantMessage content={msg.content} />
                ) : (
                  <UserMessage content={msg.content} />
                )}
              </div>
            ))}
        </div>
      </div>

      {readOnly ? (
        <div className="border-t px-4 py-3 text-center text-sm text-muted-foreground">
          {t("chat.channel.readOnly")}
        </div>
      ) : (
        <ChatComposer
          input={input}
          onInputChange={setInput}
          onSend={handleSend}
          isConnected={isChatConnected}
          hasDefaultModel={Boolean(defaultModelName)}
          onSendWithAttachments={(content, attachments) => {
            void handleSendWithAttachments(content, attachments)
          }}
          chatMode={chatMode}
          onCycleMode={cycleChatMode}
        />
      )}
    </div>
  )
}
