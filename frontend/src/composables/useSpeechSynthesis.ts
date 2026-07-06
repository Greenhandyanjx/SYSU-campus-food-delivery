/**
 * useSpeechSynthesis.ts — Web Speech API 语音输出
 *
 * 封装 SpeechSynthesis，提供：
 *  - 朗读/停止文本
 *  - 中文语音自动匹配
 *  - 状态检测
 */
import { ref, onUnmounted } from 'vue'

export function useSpeechSynthesis() {
  const isSpeaking = ref(false)
  const supported = ref(false)
  const currentText = ref('')

  let utterance: SpeechSynthesisUtterance | null = null

  if (typeof window !== 'undefined' && 'speechSynthesis' in window) {
    supported.value = true
  }

  /** 找一个合适的中文语音 */
  function pickChineseVoice(voices: SpeechSynthesisVoice[]): SpeechSynthesisVoice | null {
    // 偏好顺序：中文女声 → 中文男声 → 任何中文 → 默认
    const prefs = ['zh-CN', 'zh-HK', 'zh-TW', 'zh']
    for (const lang of prefs) {
      const matched = voices.filter(v => v.lang.startsWith(lang))
      if (matched.length > 0) {
        // 优先选女性语音（名称含 "Female" 或 "Xiaoxiao"/"Huihui" 等）
        const female = matched.find(v =>
          /female|xiaoxiao|huihui|yaoyao|mei|女/i.test(v.name)
        )
        return female || matched[0]
      }
    }
    return voices.find(v => v.lang.startsWith('zh')) || voices[0] || null
  }

  function speak(text: string) {
    if (!supported.value || !text.trim()) return

    // 如果正在朗读同一段，停止
    if (isSpeaking.value && currentText.value === text) {
      stop()
      return
    }

    // 停止当前朗读
    window.speechSynthesis.cancel()

    currentText.value = text
    utterance = new SpeechSynthesisUtterance(text)
    utterance.lang = 'zh-CN'
    utterance.rate = 1.1   // 稍快一点
    utterance.pitch = 1.0
    utterance.volume = 1.0

    // 语音就绪后设置中文语音
    const voices = window.speechSynthesis.getVoices()
    const chineseVoice = pickChineseVoice(voices)
    if (chineseVoice) {
      utterance.voice = chineseVoice
    } else {
      // voices 可能还没加载完，等 onvoiceschanged
      const onChanged = () => {
        const v = pickChineseVoice(window.speechSynthesis.getVoices())
        if (v && utterance) utterance.voice = v
        window.speechSynthesis.removeEventListener('voiceschanged', onChanged)
      }
      window.speechSynthesis.addEventListener('voiceschanged', onChanged)
    }

    utterance.onstart = () => { isSpeaking.value = true }
    utterance.onend = () => { isSpeaking.value = false; currentText.value = '' }
    utterance.onerror = () => { isSpeaking.value = false; currentText.value = '' }

    window.speechSynthesis.speak(utterance)
  }

  function stop() {
    window.speechSynthesis.cancel()
    isSpeaking.value = false
    currentText.value = ''
    utterance = null
  }

  function toggle(text: string) {
    if (isSpeaking.value && currentText.value === text) {
      stop()
    } else {
      speak(text)
    }
  }

  onUnmounted(() => {
    stop()
  })

  return { isSpeaking, supported, currentText, speak, stop, toggle }
}
