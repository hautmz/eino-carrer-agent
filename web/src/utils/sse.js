export function chatSSE(url, body, callbacks = {}) {
  const token = localStorage.getItem('eino_career_token')
  const controller = new AbortController()

  fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: token ? `Bearer ${token}` : '',
    },
    body: JSON.stringify(body),
    signal: controller.signal,
  })
    .then((response) => {
      if (!response.ok) {
        const errMsg = `HTTP ${response.status}`
        if (callbacks.onError) callbacks.onError(errMsg)
        return
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      function read() {
        reader
          .read()
          .then(({ done, value }) => {
            if (done) {
              if (callbacks.onDone) callbacks.onDone()
              return
            }

            buffer += decoder.decode(value, { stream: true })
            const parts = buffer.split('\n\n')
            buffer = parts.pop()

            for (const part of parts) {
              if (!part.trim()) continue

              let currentEvent = ''
              let currentData = ''

              const lines = part.split('\n')
              for (const line of lines) {
                if (line.startsWith('event:')) {
                  currentEvent = line.slice(6).trim()
                } else if (line.startsWith('data:')) {
                  currentData = line.slice(5).trim()
                }
              }

              if (currentEvent === 'heartbeat') continue
              if (currentEvent === 'done') {
                if (callbacks.onDone) callbacks.onDone()
                return
              }

              if (callbacks.onEvent) {
                callbacks.onEvent(currentEvent, currentData)
              }

              if (currentEvent === 'message' && callbacks.onMessage) {
                try {
                  const parsed = JSON.parse(currentData)
                  callbacks.onMessage(parsed.content || '')
                } catch {
                  callbacks.onMessage(currentData)
                }
              }
              if (currentEvent === 'tool_call' && callbacks.onToolCall) {
                try { callbacks.onToolCall(JSON.parse(currentData)) } catch { callbacks.onToolCall(currentData) }
              }
              if (currentEvent === 'report_progress' && callbacks.onReportProgress) {
                try { callbacks.onReportProgress(JSON.parse(currentData)) } catch { callbacks.onReportProgress(currentData) }
              }
              if (currentEvent === 'report_result' && callbacks.onReportResult) {
                try { callbacks.onReportResult(JSON.parse(currentData)) } catch { callbacks.onReportResult(currentData) }
              }
              if (currentEvent === 'error' && callbacks.onError) {
                callbacks.onError(currentData)
              }
            }

            read()
          })
          .catch((err) => {
            if (err.name !== 'AbortError' && callbacks.onError) {
              callbacks.onError(err.message)
            }
          })
      }

      read()
    })
    .catch((err) => {
      if (err.name !== 'AbortError' && callbacks.onError) {
        callbacks.onError(err.message)
      }
    })

  return controller
}
