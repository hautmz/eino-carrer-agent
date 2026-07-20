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
        throw new Error(`HTTP ${response.status}`)
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
            const lines = buffer.split('\n')
            buffer = lines.pop()

            let currentEvent = ''
            for (const line of lines) {
              if (line.startsWith('event: ')) {
                currentEvent = line.slice(7).trim()
              } else if (line.startsWith('data: ')) {
                const data = line.slice(6)
                if (currentEvent === 'heartbeat') continue
                if (currentEvent === 'done') {
                  if (callbacks.onDone) callbacks.onDone()
                  return
                }
                if (callbacks.onEvent) {
                  callbacks.onEvent(currentEvent, data)
                }
                if (currentEvent === 'message' && callbacks.onMessage) {
                  try {
                    const parsed = JSON.parse(data)
                    callbacks.onMessage(parsed.content || '')
                  } catch {
                    callbacks.onMessage(data)
                  }
                }
                if (currentEvent === 'tool_call' && callbacks.onToolCall) {
                  try {
                    callbacks.onToolCall(JSON.parse(data))
                  } catch {
                    callbacks.onToolCall(data)
                  }
                }
                if (currentEvent === 'report_progress' && callbacks.onReportProgress) {
                  try {
                    callbacks.onReportProgress(JSON.parse(data))
                  } catch {
                    callbacks.onReportProgress(data)
                  }
                }
                if (currentEvent === 'report_result' && callbacks.onReportResult) {
                  try {
                    callbacks.onReportResult(JSON.parse(data))
                  } catch {
                    callbacks.onReportResult(data)
                  }
                }
                if (currentEvent === 'error' && callbacks.onError) {
                  callbacks.onError(data)
                }
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
