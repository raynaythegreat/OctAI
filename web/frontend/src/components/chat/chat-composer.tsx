import { IconArrowUp, IconMicrophone } from "@tabler/icons-react"
import type { KeyboardEvent } from "react"
import * as React from "react"
import { useTranslation } from "react-i18next"
import TextareaAutosize from "react-textarea-autosize"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

interface ChatComposerProps {
  input: string
  onInputChange: (value: string) => void
  onSend: () => void
  isConnected: boolean
  hasDefaultModel: boolean
}

// Minimal Web Speech API types (not in all TS DOM lib versions)
type SpeechRecognitionCtor = new () => {
  continuous: boolean
  interimResults: boolean
  onstart: (() => void) | null
  onend: (() => void) | null
  onerror: (() => void) | null
  onresult:
    | ((ev: {
        resultIndex: number
        results: { length: number; [i: number]: { [0]: { transcript: string } } }
      }) => void)
    | null
  start(): void
  stop(): void
}

declare global {
  interface Window {
    SpeechRecognition?: SpeechRecognitionCtor
    webkitSpeechRecognition?: SpeechRecognitionCtor
  }
}

export function ChatComposer({
  input,
  onInputChange,
  onSend,
  isConnected,
  hasDefaultModel,
}: ChatComposerProps) {
  const { t } = useTranslation()
  const canInput = isConnected && hasDefaultModel
  const recognitionRef = React.useRef<InstanceType<SpeechRecognitionCtor> | null>(null)
  const [isListening, setIsListening] = React.useState(false)
  const [speechSupported] = React.useState(
    () =>
      typeof window !== "undefined" &&
      !!(window.SpeechRecognition || window.webkitSpeechRecognition),
  )

  const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.nativeEvent.isComposing) return
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      onSend()
    }
  }

  const toggleListening = () => {
    if (isListening) {
      recognitionRef.current?.stop()
      return
    }
    const SR = window.SpeechRecognition ?? window.webkitSpeechRecognition
    if (!SR) return
    const rec = new SR()
    rec.continuous = false
    rec.interimResults = true
    rec.onstart = () => setIsListening(true)
    rec.onend = () => setIsListening(false)
    rec.onerror = () => setIsListening(false)
    rec.onresult = (ev) => {
      let transcript = ""
      for (let i = ev.resultIndex; i < ev.results.length; i++) {
        transcript += ev.results[i][0].transcript
      }
      onInputChange(transcript)
    }
    recognitionRef.current = rec
    rec.start()
  }

  return (
    <div className="bg-background shrink-0 px-4 pt-4 pb-[calc(1rem+env(safe-area-inset-bottom))] md:px-8 md:pb-8 lg:px-24 xl:px-48">
      <div className="bg-card border-border/80 mx-auto flex max-w-[1000px] flex-col rounded-2xl border p-3 shadow-md">
        <TextareaAutosize
          value={input}
          onChange={(e) => onInputChange(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={t("chat.placeholder")}
          disabled={!canInput}
          className={cn(
            "placeholder:text-muted-foreground max-h-[200px] min-h-[60px] resize-none border-0 bg-transparent px-2 py-1 text-[15px] shadow-none transition-colors focus-visible:ring-0 focus-visible:outline-none dark:bg-transparent",
            !canInput && "cursor-not-allowed",
          )}
          minRows={1}
          maxRows={8}
        />

        <div className="mt-2 flex items-center justify-between px-1">
          <div className="flex items-center gap-1">
            {speechSupported && (
              <Button
                type="button"
                size="icon"
                variant="ghost"
                className={cn(
                  "size-8 rounded-full transition-all",
                  isListening
                    ? "bg-violet-500/20 text-violet-400 animate-pulse"
                    : "text-muted-foreground hover:text-violet-400 hover:bg-violet-500/10",
                )}
                onClick={toggleListening}
                disabled={!canInput}
                title={isListening ? "Stop recording" : "Voice input"}
              >
                <IconMicrophone className="size-4" />
              </Button>
            )}
          </div>

          <Button
            size="icon"
            className="size-8 rounded-full bg-violet-500 text-white transition-transform hover:bg-violet-600 active:scale-95"
            onClick={onSend}
            disabled={!input.trim() || !canInput}
          >
            <IconArrowUp className="size-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
