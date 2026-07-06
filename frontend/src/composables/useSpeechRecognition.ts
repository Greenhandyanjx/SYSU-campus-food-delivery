/**
 * useSpeechRecognition.ts — Web Speech API 语音输入
 *
 * 封装 SpeechRecognition，提供：
 *  - 开始/停止录音
 *  - 实时转写（interim results）
 *  - 最终识别文本回调
 *  - 浏览器兼容检测
 *
 * 使用方式：
 *   const { isListening, transcript, supported, start, stop } = useSpeechRecognition({
 *     onResult: (text) => inputMsg.value = text,
 *     onEnd: () => onSend(),
 *   })
 */
import { ref, onUnmounted } from 'vue'

export interface SpeechRecognitionOptions {
  /** 每次有识别结果时回调（interim + final） */
  onResult?: (text: string, isFinal: boolean) => void
  /** 一次识别会话完全结束时回调 */
  onEnd?: () => void
  /** 发生错误时回调 */
  onError?: (error: string) => void
  /** 语言（默认 zh-CN） */
  lang?: string
}

export function useSpeechRecognition(opts: SpeechRecognitionOptions = {}) {
  const isListening = ref(false)
  const transcript = ref('')
  const error = ref('')
  const supported = ref(false)

  let recognition: any = null
  let manualStop = false

  // 检测浏览器兼容性
  const SpeechRecognitionCtor =
    (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition
  if (SpeechRecognitionCtor) {
    supported.value = true
  }

  function start() {
    if (!SpeechRecognitionCtor) {
      error.value = '浏览器不支持语音识别'
      opts.onError?.('浏览器不支持语音识别')
      return
    }

    if (isListening.value) return

    try {
      manualStop = false
      recognition = new SpeechRecognitionCtor()
      recognition.lang = opts.lang || 'zh-CN'
      recognition.continuous = false   // 一段话说完即停止
      recognition.interimResults = true // 实时显示中间结果
      recognition.maxAlternatives = 1

      recognition.onresult = (event: SpeechRecognitionEvent) => {
        let interim = ''
        let final = ''
        for (let i = event.resultIndex; i < event.results.length; i++) {
          const isFinal = event.results[i].isFinal
          const text = event.results[i][0].transcript
          if (isFinal) {
            final += text
          } else {
            interim += text
          }
        }
        const text = final || interim
        transcript.value = text
        if (final || interim) {
          opts.onResult?.(text, !!final && !interim)
        }
      }

      recognition.onend = () => {
        isListening.value = false
        if (!manualStop) {
          opts.onEnd?.()
        }
      }

      recognition.onerror = (event: any) => {
        const msg = event.error || '未知错误'
        error.value = msg
        isListening.value = false
        opts.onError?.(msg)
      }

      recognition.start()
      isListening.value = true
      error.value = ''
    } catch (e: any) {
      error.value = e.message || '启动语音识别失败'
      isListening.value = false
      opts.onError?.(error.value)
    }
  }

  function stop() {
    manualStop = true
    if (recognition && isListening.value) {
      try { recognition.stop() } catch { /* ignore */ }
    }
    isListening.value = false
  }

  function toggle() {
    if (isListening.value) {
      stop()
    } else {
      start()
    }
  }

  onUnmounted(() => {
    stop()
    recognition = null
  })

  return { isListening, transcript, error, supported, start, stop, toggle }
}
